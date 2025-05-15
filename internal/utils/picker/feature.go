package picker

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/components/selector"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/grouping"
	"github.com/spf13/cobra"
	"slices"
	"strings"
)

const (
	nowSection   = "now-projects"
	nextSection  = "next-projects"
	laterSection = "later-projects"
)

type TargetCmd struct {
	CompanyID  int32
	ProjectID  string
	FeatureID  string
	IDEName    string
	Yes        bool
	SingleShot bool
}

func (c *TargetCmd) PreRun(_ *cobra.Command, _ []string) error {
	allowedIDEs := ide.GetKnown()
	// Validate mode flag if provided
	if c.IDEName != "" && !slices.Contains(allowedIDEs, ide.IDE(c.IDEName)) {
		return fmt.Errorf("allowed values for IDE are %v, got: %s", allowedIDEs, c.IDEName)
	}
	if c.SingleShot && c.FeatureID != "" {
		return fmt.Errorf("-s (--single-shot) cannot be used with -f (--feature)")
	}
	return nil
}

func (c *TargetCmd) Prepare(cmd *cobra.Command) {
	knownIDEs := ide.GetKnown()
	var ideStr []string
	for _, i := range knownIDEs {
		ideStr = append(ideStr, string(i))
	}
	cmd.Flags().StringVarP(
		&c.IDEName, "ide", "i", "", fmt.Sprintf("IDE to use. Allowed values: %v", strings.Join(ideStr, ", ")))
	cmd.Flags().StringVarP(&c.ProjectID, "project", "p", "", "Project id to focus on")
	cmd.Flags().StringVarP(&c.FeatureID, "feature", "f", "", "Target id to focus on")
	cmd.Flags().Int32VarP(&c.CompanyID, "company", "c", -1, "Company id to focus on")
	cmd.Flags().BoolVarP(&c.Yes, "yes", "y", false, "Execute without confirmation")
	cmd.Flags().BoolVarP(&c.SingleShot, "single-shot", "s", false, "Use a single-shot prompt for all features (cannot be used with -f)")
}

type DevTarget struct {
	SpecificFeature *documents.DocumentEntity
	SingleShot      bool
	ProjectWithDocs *documents.ProjectWithDocs
}

func (d DevTarget) GetName() string {
	if d.SingleShot {
		return d.ProjectWithDocs.GetProject().GetTitle()
	}
	return d.SpecificFeature.GetTitle()
}

func mainGroupID(companyID int32) string {
	return fmt.Sprintf("%v-projects", companyID)
}

func Target(cmd *TargetCmd) (DevTarget, error) {
	cl := devplan.NewClient(devplan.Config{})
	self, err := cl.GetSelf()
	if err != nil {
		return DevTarget{}, err
	}
	companyID := cmd.CompanyID
	projectID := cmd.ProjectID
	featureID := cmd.FeatureID
	singleShot := cmd.SingleShot
	companies := self.GetOwnInfo().GetCompanyDetails()
	company, err := selector.Company(companies, selector.Props{}, companyID)
	if err != nil {
		return DevTarget{}, err
	}
	project, err := selectProject(cl, company.GetId(), projectID)
	if err != nil {
		return DevTarget{}, err
	}
	result := DevTarget{
		SingleShot:      singleShot,
		ProjectWithDocs: project,
	}
	if !result.SingleShot {
		features := getFeatures(project)
		feature, err := selector.Feature(features, selector.Props{}, featureID)
		if err != nil {
			return DevTarget{}, err
		}
		result.SpecificFeature = feature
	}

	return result, nil
}

func selectProject(cl *devplan.Client, companyID int32, projectID string) (*documents.ProjectWithDocs, error) {
	prResp, err := cl.GetCompanyProjects(companyID)
	if err != nil {
		return nil, err
	}

	if projectID != "" {
		for _, p := range prResp.GetProjects() {
			if p.GetProject().GetId() == projectID {
				return getProjectWithDocs(projectID, prResp.GetProjects()), nil
			}
		}
		return nil, fmt.Errorf("project with id %v not found", projectID)
	}
	grResp, err := cl.GetGroup(companyID, mainGroupID(companyID))
	if err != nil {
		return nil, err
	}

	var projects []*documents.ProjectEntity
	for _, p := range prResp.GetProjects() {
		if p.GetProject().GetDetails().GetStatus() != documents.ProjectStatus_DONE {
			projects = append(projects, p.GetProject())
		}
	}
	ordered := orderedProjects(grResp.GetGroup(), projects)
	selectedPr, err := selector.Project(ordered, selector.Props{})
	if err != nil {
		return nil, err
	}
	return getProjectWithDocs(selectedPr.GetId(), prResp.GetProjects()), nil
}

func getFeatures(project *documents.ProjectWithDocs) []*documents.DocumentEntity {
	var docs []*documents.DocumentEntity
	for _, d := range project.GetDocs() {
		if d.GetType() == documents.DocumentType_FEATURE {
			docs = append(docs, d)
		}
	}
	return docs
}

func orderedProjects(group *grouping.Group, projects []*documents.ProjectEntity) []*documents.ProjectEntity {
	result := getProjectsFromSections(projects, group)
	known := make(map[string]bool)
	for _, p := range result {
		known[p.GetId()] = true
	}
	for _, p := range projects {
		if !known[p.GetId()] {
			result = append(result, p)
		}
	}
	return result
}

func getProjectsFromSections(projects []*documents.ProjectEntity, group *grouping.Group) []*documents.ProjectEntity {
	var result []*documents.ProjectEntity
	for _, k := range []string{nowSection, nextSection, laterSection} {
		section := getSection(group, k)
		pr := getSectionProjects(projects, section)
		result = append(result, pr...)
	}
	return result
}

func getSectionProjects(projects []*documents.ProjectEntity, section *grouping.GroupItem) []*documents.ProjectEntity {
	if section == nil {
		return nil
	}
	byID := make(map[string]*documents.ProjectEntity)
	for _, p := range projects {
		byID[p.GetId()] = p
	}
	var result []*documents.ProjectEntity
	for _, itemID := range section.GetChildItems() {
		if p, ok := byID[itemID]; ok {
			result = append(result, p)
		}
	}
	return result
}

func getSection(group *grouping.Group, key string) *grouping.GroupItem {
	for _, item := range group.GetItems() {
		if item.GetKey() == key {
			if item.GetSection() != nil {
				return item
			}
			break
		}
	}
	return nil
}

func getProjectWithDocs(projectID string, projects []*documents.ProjectWithDocs) *documents.ProjectWithDocs {
	for _, p := range projects {
		if p.GetProject().GetId() == projectID {
			return p
		}
	}
	return nil
}
