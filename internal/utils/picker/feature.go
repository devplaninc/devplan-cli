package picker

import (
	"fmt"
	"slices"
	"strings"

	"github.com/devplaninc/devplan-cli/internal/components/selector"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/grouping"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/templates"
	"github.com/spf13/cobra"
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
	TaskID     string
	IDEName    string
	Yes        bool
	SingleShot bool
	Steps      bool
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
	if c.Steps && c.FeatureID == "" {
		return fmt.Errorf("--steps must be used together with -f (--feature)")
	}
	if c.TaskID != "" && c.SingleShot {
		return fmt.Errorf("--task cannot be used with -s (--single-shot)")
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
	cmd.Flags().StringVar(&c.TaskID, "task", "", "Task id to focus on (optional)")
	cmd.Flags().Int32VarP(&c.CompanyID, "company", "c", -1, "Company id to focus on")
	cmd.Flags().BoolVarP(&c.Yes, "yes", "y", false, "Execute without confirmation")
	cmd.Flags().BoolVarP(&c.SingleShot, "single-shot", "s", false, "Use a single-shot prompt for all features (cannot be used with -f)")
	cmd.Flags().BoolVar(&c.Steps, "steps", false, "Use step-by-step prompts for a feature")
}

type DevTarget struct {
	SpecificFeature *documents.DocumentEntity
	SingleShot      bool
	ProjectWithDocs *documents.ProjectWithDocs
	Steps           bool
	Task            *documents.DocumentEntity
	Template        *templates.ProjectTemplate
}

func (d DevTarget) GetName() string {
	if t := d.Task; t != nil {
		return t.GetTitle()
	}
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
	taskID := cmd.TaskID
	companies := self.GetOwnInfo().GetCompanyDetails()
	company, err := selector.Company(companies, selector.Props{}, companyID)
	if err != nil {
		return DevTarget{}, err
	}
	project, err := selectProject(cl, company.GetId(), projectID)
	if err != nil {
		return DevTarget{}, err
	}
	templatesRep, err := cl.GetProjectTemplates(company.GetId())
	if err != nil {
		return DevTarget{}, err
	}
	result := DevTarget{
		SingleShot:      singleShot,
		ProjectWithDocs: project,
	}
	for _, t := range templatesRep.GetProjectTemplates() {
		if t.GetId() == project.GetProject().GetTemplateId() {
			result.Template = t
			break
		}
	}
	features := getFeatures(project)
	tasks := getTasks(project)
	if taskID != "" {
		for _, d := range tasks {
			if d.GetType() == documents.DocumentType_TASK && d.GetId() == taskID {
				result.Task = d
				break
			}
		}
	}
	if result.Task == nil {
		// First, if tasks exist at all, select a task. Only if no tasks, select feature.
		if len(tasks) > 0 {
			task, err := selector.Task(tasks, selector.Props{}, "")
			if err != nil {
				return DevTarget{}, err
			}
			result.Task = task
		} else if !result.SingleShot {
			feature, err := selector.Feature(features, selector.Props{}, featureID)
			if err != nil {
				return DevTarget{}, err
			}
			result.SpecificFeature = feature
			result.Steps = cmd.Steps
		}
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
	return getTypedDocs(project, documents.DocumentType_FEATURE)
}

func getTasks(project *documents.ProjectWithDocs) []*documents.DocumentEntity {
	return getTypedDocs(project, documents.DocumentType_TASK)
}

func getTypedDocs(project *documents.ProjectWithDocs, docType documents.DocumentType) []*documents.DocumentEntity {
	var docs []*documents.DocumentEntity
	for _, d := range project.GetDocs() {
		if d.GetType() == docType {
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
