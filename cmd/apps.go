package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/cnupp/cli/config"
	"github.com/cnupp/cli/pkg"
	"github.com/cnupp/appssdk/api"
	"github.com/cnupp/appssdk/net"
	launcherApi "github.com/cnupp/runtimesdk/api"
	deploymentNet "github.com/cnupp/runtimesdk/net"
)

func askForOverrideExistingApp() bool {
	reader := bufio.NewReader(os.Stdin)
	for true {
		fmt.Printf("Another app is using this repository, are you sure to continue (y/N)?")
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(text) == "y" {
			return true
		} else if strings.TrimSpace(text) == "N" {
			return false
		}
	}

	return false
}

func AppLaunch(appId string) error {
	configRepository, appId, err := load(appId)

	if err != nil {
		return errors.New("Please execute 'cde apps:launch' inside a project with an application created for it or specify the app to be launched")
	}

	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	app, err := appRepository.GetApp(appId)
	if err != nil {
		return err
	}
	stack, err := app.GetStack()
	if err != nil {
		return err
	}

	if stack.Type() != "NON_BUILD_STACK" {
		return errors.New("only non build stack app can be launched")
	}

	releaseMapper := api.NewReleaseMapper(configRepository, net.NewCloudControllerGateway(configRepository))
	_, err = releaseMapper.Create(app)
	if err != nil {
		return err
	}
	fmt.Printf("create %s release successfully\n", app.Name())
	return nil
}

// AppCreate creates an app.
func AppCreate(appId string, stackName string, unifiedProcedure, providerName, owner string, needDeploy string) error {
	var needDeployBool bool

	if !git.IsGitDirectory() {
		return fmt.Errorf("Not in a git repository")
	}

	configRepository := config.NewConfigRepository(func(error) {})
	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))

	stackRepo := api.NewStackRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))

	appName, _ := git.DetectAppName(configRepository.GitHost())
	if appName != "" {
		if !askForOverrideExistingApp() {
			return fmt.Errorf("Give up to override existing app")
		}
	}
	var stackId string
	var stack api.Stack
	var unifiedProcedureId string
	var providerLink api.Link

	if len(stackName) != 0 {
		stacks, err := stackRepo.GetStackByName(stackName)
		if err != nil {
			return err
		}

		stack = stacks.Items()[0]

		stackId = stack.Id()
	} else {
		gateway := deploymentNet.NewCloudControllerGateway(configRepository)
		ups := launcherApi.NewUpsRepository(configRepository, gateway)
		providers := launcherApi.NewProviderRepository(configRepository, gateway)

		unifiedProcedures, err := ups.GetUPByName(unifiedProcedure)
		if err != nil {
			return err
		}

		if unifiedProcedures.Count() != 1 {
			return fmt.Errorf("can not find the unified procedure by name given")
		}

		unifiedProcedureId = unifiedProcedures.Items()[0].Id()

		provider, err := providers.GetProviderByName(providerName)
		if err != nil {
			return err
		}
		providerLink, err = provider.Links().Link("self")
		if err != nil {
			return err
		}

	}

	if needDeploy == "1" {
		needDeployBool = true
	} else {
		needDeployBool = false
	}

	appParams := api.AppParams{
		Name:             appId,
		Stack:            stackId,
		Provider:         providerLink.URI,
		Owner:            owner,
		UnifiedProcedure: unifiedProcedureId,
		NeedDeploy:       needDeployBool,
	}
	createdApp, err := appRepository.Create(appParams)
	if err != nil {
		return err
	}
	u, err := url.Parse(configRepository.ApiEndpoint())
	if err != nil {
		return err
	}
	host := u.Host
	if strings.Index(host, ":") != -1 {
		splits := strings.Split(host, ":")
		host = splits[0]
	}
	git.DeleteCdeRemote()
	host = configRepository.GitHost()
	if err = git.CreateRemote(host, "cde", createdApp.Name()); err != nil {
		if err.Error() == "exit status 128" {
			fmt.Println("To replace the existing git remote entry, run:")
			fmt.Printf("  git remote rename cde cde.old && cde git:remote -a %s\n", createdApp.Name())
		}
		return err
	}

	fmt.Println("remote available at", git.RemoteURL(host, createdApp.Name()))

	if stack != nil && stack.Type() == "NON_BUILD_STACK" {
		releaseMapper := api.NewReleaseMapper(configRepository, net.NewCloudControllerGateway(configRepository))
		_, err = releaseMapper.Create(createdApp)
		if err != nil {
			fmt.Println("error in create release for non build stack")
		} else {
			fmt.Println("create new release for non build stack")
		}
	}

	return err
}

func AppsList() error {
	configRepository := config.NewConfigRepository(func(error) {})
	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	apps, err := appRepository.GetApps()
	if err != nil {
		return err
	}
	fmt.Printf("=== Apps [%d]\n", len(apps.Items()))

	for _, app := range apps.Items() {
		fmt.Printf("id: %s\n", app.Name())
	}
	return nil
}

func GetApp(appId string) error {
	configRepository, appId, err := load(appId)
	if err != nil {
		return err
	}
	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	app, err := appRepository.GetApp(appId)
	if err != nil {
		return err
	}
	outputDescription(app)
	outputRoutes(app)
	outputDependentServices(appId)

	return nil
}

func outputDescription(app api.App) {
	fmt.Printf("--- %s Application\n", app.Name())
	data := make([][]string, 2)
	data[0] = []string{"ID", app.Name()}
	stack, _ := app.GetStack()
	data[1] = []string{"Stack Name", stack.Name()}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
}

func outputRoutes(app api.App) {
	boundRoutes, err := app.GetRoutes()
	fmt.Println("--- Access routes:\n")

	if err != nil {
		fmt.Print(err)
		return
	}
	for boundRoutes != nil {
		routes := boundRoutes.Items()
		for _, route := range routes {
			fmt.Println(route.DomainField.Name + "/" + route.PathField + " \n")
		}
		boundRoutes, _ = boundRoutes.Next()
	}

}

func outputDependentServices(appId string) error {
	configRepository, appId, err := load(appId)
	if err != nil {
		return err
	}
	repo := launcherApi.NewDeploymentRepository(configRepository, deploymentNet.NewCloudControllerGateway(configRepository))
	servicesModel, err := repo.GetDependentServicesForApp(appId)
	fmt.Print("--- Dependent services:\n")
	if err != nil {
		fmt.Print(err)
		return err
	}
	servicesArray := servicesModel
	for index, service := range servicesArray {
		fmt.Printf("-----> Service %d:\n", index+1)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Append([]string{"ID", service.Id()})
		table.Append([]string{"Name", service.Name()})
		table.Append([]string{"Instances", fmt.Sprintf("%d", service.Instance())})
		table.Append([]string{"Memory", fmt.Sprintf("%v", service.Memory())})
		table.Append([]string{"Env", service.Env()})
		table.Render() // Send output
	}
	return nil
}

func DestroyApp(appId string) error {
	configRepository, appId, err := load(appId)

	if err != nil {
		return errors.New("Please execute 'cde apps:destroy' inside a project with an application created for it or specify the app to be destroyed")
	}

	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	app, err := appRepository.GetApp(appId)
	if err != nil {
		return err
	}

	stack, err := app.GetStack()
	if err != nil {
		return err
	}

	deployRepo := launcherApi.NewDeploymentRepository(configRepository, deploymentNet.NewCloudControllerGateway(configRepository))
	err = deployRepo.Destroy(appId)
	if err != nil {
		fmt.Printf("failed to destroy %s deployment\n", app.Name())
	}

	err = appRepository.Delete(appId)
	if err != nil {
		return errors.New("No application for this project")
	}

	fmt.Printf("destroy %s successfully!\n", appId)

	if stack.Type() == "BUILD_STACK" {
		if git.HasRemoteNameForApp("cde", appId) {
			err = git.DeleteRemote(appId)
			if err != nil {
				fmt.Print("Remove 'cde' remote failed. \n Please execute git cmd in the app directory: `git remote remove cde`")
			}
		} else {
			fmt.Print("Please execute git cmd in the app directory: `git remote remove cde`")
		}
	}

	return nil
}

func SwitchStack(appName string, stackName string) error {
	configRepository, appName, err := load(appName)
	if err != nil {
		return err
	}
	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	params := api.UpdateStackParams{
		Stack: stackName,
	}
	err = appRepository.SwitchStack(appName, params)
	if err != nil {
		return err
	}
	fmt.Printf("Switch to stack '%s' successfully.\n", stackName)
	return nil
}

func AppLog(appId string, lines int) error {
	configRepository, appId, err := load(appId)
	if err != nil {
		return err
	}
	deploymentRepository := launcherApi.NewDeploymentRepository(configRepository,
		deploymentNet.NewCloudControllerGateway(configRepository))
	deployment, err := deploymentRepository.GetDeploymentByAppName(appId)
	if err != nil {
		return err
	}
	output, err := deployment.Log(lines)
	if err != nil {
		return err
	}

	if output.ErrorField != "" {
		return fmt.Errorf(output.ErrorField)
	}

	handleOutput(output)
	return nil
}

func ServiceLog(appId, serviceName string, lines int) error {
	configRepository, appId, err := load(appId)
	if err != nil {
		return err
	}
	deploymentRepository := launcherApi.NewDeploymentRepository(configRepository,
		deploymentNet.NewCloudControllerGateway(configRepository))
	deployment, err := deploymentRepository.GetDeploymentByAppName(appId)
	if err != nil {
		return err
	}

	service, err := deployment.GetService(serviceName)
	if err != nil {
		return err
	}
	output, err := service.Log(lines)

	if err != nil {
		return err
	}

	if output.ErrorField != "" {
		return fmt.Errorf(output.ErrorField)
	}

	handleOutput(output)
	return nil
}

func AppLocalization(appName string, directory string) error {
	configRepository := config.NewConfigRepository(func(error) {})
	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	_, err := appRepository.GetApp(appName)
	if err != nil {
		return err
	}

	host := configRepository.GitHost()
	gitRepo := git.RemoteURL(host, appName)
	if directory == "" {
		directory = appName
	}

	currentDir, _ := os.Getwd()
	target := fmt.Sprintf("%s//%s", currentDir, directory)

	if IsDirectoryExist(target) {
		return fmt.Errorf("directory %s already exists", directory)
	}

	if err = git.CloneRepo(gitRepo, directory); err != nil {
		return err
	}

	if err = os.Chdir(target); err != nil {
		return err
	}
	git.DeleteCdeRemote()

	if err = git.CreateRemote(host, "cde", appName); err != nil {
		if err.Error() == "exit status 128" {
			fmt.Println("To replace the existing git remote entry, run:")
			fmt.Printf("  git remote rename cde cde.old && cde git:remote -a %s\n", appName)
		}
		return err
	}
	return nil
}

func handleOutput(output api.LogsModel) {
	for _, log := range output.ItemsField {
		fmt.Printf("%s\n", log.MessageField)
	}
}

func AppCollaborators(appId string) error {
	configRepository, appId, err := load(appId)

	if err != nil {
		return errors.New("Please execute 'cde apps:collaborators' inside a project with an application created for it or specify the app")
	}

	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	app, err := appRepository.GetApp(appId)
	if err != nil {
		return err
	}

	users, err := app.GetCollaborators()
	if err != nil {
		return err
	}

	fmt.Printf("=== Collaborators: [%d]\n", len(users))

	for _, user := range users {
		fmt.Printf("email: %s\n", user.Email())
	}
	return nil
}

func AppAddCollaborator(appId string, email string) error {
	configRepository, appId, err := load(appId)

	if err != nil {
		return errors.New("Please execute 'cde apps:add-collaborators' inside a project with an application created for it or specify the app")
	}

	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	app, err := appRepository.GetApp(appId)
	if err != nil {
		return err
	}

	err = app.AddCollaborator(api.CreateCollaboratorParams{
		Email: email,
	})
	if err != nil {
		return err
	}
	fmt.Print("Add collaborator success.\n")
	return nil
}

func AppRmCollaborator(appId string, email string) error {
	configRepository, appId, err := load(appId)

	if err != nil {
		return errors.New("Please execute 'cde apps:rm-collaborators' inside a project with an application created for it or specify the app")
	}

	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	app, err := appRepository.GetApp(appId)
	if err != nil {
		return err
	}

	userRepository := api.NewUserRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	Users, err := userRepository.GetUserByEmail(email)
	if err != nil {
		return err
	}

	if len(Users.Items()) <= 0 {
		return errors.New(fmt.Sprintf("no such user %s", email))
	}

	user := Users.Items()[0]

	err = app.RemoveCollaborator(user.Id())
	if err != nil {
		return err
	}
	fmt.Print("Remove collaborator success.\n")
	return nil
}

func AppTransfer(appId string, email string, org string) error {
	if email == "" && org == "" {
		return errors.New(fmt.Sprint("Email or Org Name should given."))
	}
	if email != "" && org != "" {
		return errors.New(fmt.Sprint("Only one of Email and Org Name should given."))
	}

	configRepository, appId, err := load(appId)

	if err != nil {
		return errors.New("Please execute 'cde apps:transfer' inside a project with an application created for it or specify the app")
	}

	appRepository := api.NewAppRepository(configRepository,
		net.NewCloudControllerGateway(configRepository))
	app, err := appRepository.GetApp(appId)
	if err != nil {
		return err
	}

	if email != "" {
		err = app.TransferToUser(email)
		if err != nil {
			return err
		}
		fmt.Printf("Finish transfer %s to %s\n", app.Name(), email)
	} else {
		err = app.TransferToOrg(org)
		if err != nil {
			return err
		}
		fmt.Printf("Finish transfer %s to %s\n", app.Name(), org)
	}

	return nil
}

func IsAppNameInvalid(appName string) bool {
	regex := regexp.MustCompile(`^[a-z0-9\-]+$`)

	return regex.MatchString(appName)
}
