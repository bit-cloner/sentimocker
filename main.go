package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/hashicorp/go-slug"
	tfe "github.com/hashicorp/go-tfe"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func after(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}
func main() {
	tftoken := ""
	prompt := &survey.Password{
		Message: "Please type your TF enterprise API token",
	}
	survey.AskOne(prompt, &tftoken)

	config := &tfe.Config{
		Token: tftoken,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	orgs, err := client.Organizations.List(context.Background(), tfe.OrganizationListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%d\n", orgs.Items)
	// store organisation values as an array of string
	var orglist []string
	for _, orgitem := range orgs.Items {
		orglist = append(orglist, orgitem.Name)
	}

	selectedorg := ""
	selection := &survey.Select{
		Message: "Choose an Organization:",
		Options: orglist,
	}
	survey.AskOne(selection, &selectedorg)
	// get workspaces from selected organisation
	wrkspaces, err := client.Workspaces.List(context.Background(), selectedorg, tfe.WorkspaceListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%d\n", wrkspaces.Items)
	var wslist []string
	for _, wsitem := range wrkspaces.Items {
		wslist = append(wslist, wsitem.Name+"=>"+wsitem.ID)
	}

	//Ask user to choose a workspace
	selectedws := ""
	selection = &survey.Select{
		Message: "Choose a Workspace:",
		Options: wslist,
	}
	survey.AskOne(selection, &selectedws)
	// extract workspaceID from selected option
	//fmt.Println(after(selectedws, "=>"))
	selectedwsid := after(selectedws, "=>")

	// Get all runs from selected workspace
	runs, err := client.Runs.List(context.Background(), selectedwsid, tfe.RunListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%d\n", runs.Items)
	var runlist []string
	for _, runitem := range runs.Items {
		//fmt.Println(runitem.Message)
		runlist = append(runlist, runitem.Message+"=>"+runitem.Plan.ID)
	}

	//Ask user to choose a Run
	selectedrun := ""
	selection = &survey.Select{
		Message: "Choose a Run based on commit message:",
		Options: runlist,
	}
	survey.AskOne(selection, &selectedrun)

	selectedpid := after(selectedrun, "=>")
	// Read a plan base don planid to get plan exportID.
	fmt.Print("Reading Plan Export for Plan ID \n")
	plan, err := client.Plans.Read(context.Background(), selectedpid)
	if err != nil {
		log.Fatal(err, "failed to read Plan ")
	}
	fmt.Println("Found plan")
	var planExportID string
	if plan.Exports == nil {
		fmt.Print("Requesting Plan Export ... \n")
		planExport, err := client.PlanExports.Create(context.Background(), tfe.PlanExportCreateOptions{
			Plan:     plan,
			DataType: tfe.PlanExportType(tfe.PlanExportSentinelMockBundleV0),
		})
		if err != nil {
			log.Fatal(err, "failed to read Plan ")
		}
		planExportID = planExport.ID
		fmt.Println("Status of plan export is ", planExport.Status) // When Status is "finished" mocks are ready to be downloaded.
		if planExport.Status != "finished" {
			fmt.Println("waiting for Plan export to be ready for download .....⌛")
			time.Sleep(6 * time.Second)
		}
		// TODO check if export status is ready to download
	} else {
		fmt.Print("Found existing Plan Export \n")
		planExportID = plan.Exports[0].ID // Just grab the first one?
	}
	fmt.Println("Plan export ID is: ", planExportID)
	// Now download plan export by giving plan export id
	buff, err := client.PlanExports.Download(context.Background(), planExportID)
	if err != nil {
		log.Fatal(err, "failed to download plan export most likely due to export being not ready for download. try again.")
	}
	reader := bytes.NewReader(buff)
	// Save the downloaded object onto disk
	fmt.Println("Saving into directory mocks")
	dst, err := ioutil.TempDir("", "slug")
	if err != nil {
		log.Fatal(err, "failed to create directory")
	}
	//fmt.Println("temporary directory created is ", dst)
	// creating sub direcotry to store mocks
	if _, err := os.Stat("/mocks"); os.IsNotExist(err) {
		os.Mkdir("mocks", 0755)
	}
	if err != nil {
		log.Fatal(err, "failed to create  mocks directory")
	}
	// remove any old contents from mocks directory
	err = os.RemoveAll("/mocks/")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dst)
	// Unpack contents of the download
	if err := slug.Unpack(reader, "mocks"); err != nil {
		log.Fatal(err, "failed to unpack")
	}
	fmt.Println("Downloaded and unpacked into /mocks")
}
