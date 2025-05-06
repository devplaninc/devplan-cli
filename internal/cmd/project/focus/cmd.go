package focus

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/components/selector"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/devplaninc/devplan-cli/internal/utils/ide/cursor"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/grouping"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"slices"
	"strings"
)

const (
	nowSection   = "now-projects"
	nextSection  = "next-projects"
	laterSection = "later-projects"
)

var (
	Cmd = create()

	allowedIDEs = []string{"cursor"}
)

func mainGroupID(companyID int32) string {
	return fmt.Sprintf("%v-projects", companyID)
}

func create() *cobra.Command {
	var ideName string
	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Focus on a specific feature of the project",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate mode flag if provided
			if !slices.Contains(allowedIDEs, ideName) {
				return fmt.Errorf("IDE must be provided, allowed values %v, got: %s", allowedIDEs, ideName)
			}
			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			repo, err := git.CurrentRepo()
			check(err)
			fmt.Printf("Cur repo: %+v\n", repo.FullNames[0])

			cl := devplan.NewClient(devplan.ClientConfig{})
			self, err := cl.GetSelf()
			check(err)
			companies := self.GetOwnInfo().GetCompanyDetails()
			company, err := selector.Company(companies, selector.Props{})
			check(err)
			grResp, err := cl.GetGroup(company.GetId(), mainGroupID(company.GetId()))
			check(err)
			prResp, err := cl.GetCompanyProjects(company.GetId())
			check(err)
			var projects []*documents.ProjectEntity
			for _, p := range prResp.GetProjects() {
				if p.GetProject().GetDetails().GetStatus() != documents.ProjectStatus_DONE {
					projects = append(projects, p.GetProject())
				}
			}
			ordered := orderedProjects(grResp.GetGroup(), projects)
			selectedPr, err := selector.Project(ordered, selector.Props{})
			check(err)
			project := getProjectWithDocs(selectedPr.GetId(), prResp.GetProjects())
			features := getFeatures(project)
			feature, err := selector.Feature(features, selector.Props{})
			check(err)
			featPrompt, err := getFeaturePrompt(feature.GetId(), project.GetDocs())
			check(err)
			sumResp, err := cl.GetRepoSummaries(company.GetId())
			check(err)
			summary := getMatchingSummary(repo, sumResp.GetSummaries())

			if ideName == "cursor" {
				err = createCursorRules(featPrompt, summary)
				check(err)
				fmt.Println(out.Highlight("Cursor rules created successfully!"))
			}
		},
	}
	cmd.Flags().StringVarP(
		&ideName, "ide", "i", "", fmt.Sprintf("IDE to use. Allowed values: %v", strings.Join(allowedIDEs, ", ")))

	return cmd
}

func confirmRulesGeneration(
	ideName string,
	featurePrompt *documents.DocumentEntity,
	repoSummary *artifacts.ArtifactRepoSummary,
) error {
	if featurePrompt == nil && repoSummary == nil {
		return fmt.Errorf("neither feature prompt nor repo summary found for the feature and repository")
	}
	root, err := git.GetRoot()
	if err != nil {
		return err
	}
	fmt.Printf(out.Highlight(fmt.Sprintf(
		"\n%s rules will be generated for the selected feature in the current repository %v.\n\n", ideName, root)))
	prompt := promptui.Prompt{
		Label:     "Create rules",
		IsConfirm: true,
	}
	result, err := prompt.Run()
	if err != nil {
		return err
	}
	if result != "y" {
		return fmt.Errorf("aborted:" + result)
	}
	return nil
}

//func createJunieRules(featurePrompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
//	if err := confirmRulesGeneration("Junie", featurePrompt, repoSummary); err != nil {
//		return err
//	}
//}

func createCursorRules(featurePrompt *documents.DocumentEntity, repoSummary *artifacts.ArtifactRepoSummary) error {
	if err := confirmRulesGeneration("Cursor", featurePrompt, repoSummary); err != nil {
		return err
	}
	rules := []ide.Rule{
		{Name: "flow", Content: devFlowRule, Header: cursor.GetRuleHeader(cursor.Header{
			Description: devFlowRuleDescription,
		})},
		{Name: "rules", Content: rulesRule, Header: cursor.GetRuleHeader(cursor.Header{
			Description: generalRuleDescription, Globs: ".cursor/rules/*.mdc",
		})},
		{Name: "insights", Content: insightsRule, Header: cursor.GetRuleHeader(cursor.Header{
			Description: insightsRuleDescription,
		})},
	}
	if featurePrompt != nil {
		rules = append(rules, ide.Rule{
			Name:    "current_feature",
			Content: featurePrompt.GetContent(),
			Header: cursor.GetRuleHeader(cursor.Header{
				Description: currentFeatRuleDescription,
			}),
		})
	}
	if summary := repoSummary.GetSummary().GetContent(); summary != "" {
		rules = append(rules, ide.Rule{Name: "repo_overview", Content: summary, Header: cursor.GetRuleHeader(
			cursor.Header{Description: repoOverviewRuleDescription},
		)})
	}

	err := cursor.CreateRules(rules)
	if err != nil {
		return err
	}
	return nil
}

func getMatchingSummary(repo git.RepoInfo, summaries []*artifacts.ArtifactRepoSummary) *artifacts.ArtifactRepoSummary {
	for _, s := range summaries {
		if repo.MatchesName(s.GetRepoName()) {
			return s
		}
	}
	return nil
}

func getFeaturePrompt(featureID string, docs []*documents.DocumentEntity) (*documents.DocumentEntity, error) {
	codeAssist := getCodingAssistant(docs)
	if codeAssist == nil {
		return nil, nil
	}
	for _, d := range docs {
		if d.GetParentId() != codeAssist.GetId() {
			continue
		}
		details := &documents.CustomDocumentDetails{}
		err := protojson.Unmarshal([]byte(d.GetDetails()), details)
		if err != nil {
			return nil, err
		}
		if details.GetExtraPromptParams()["feature_id"] == featureID {
			return d, nil
		}
	}
	return nil, nil
}

func getCodingAssistant(docs []*documents.DocumentEntity) *documents.DocumentEntity {
	for _, d := range docs {
		if d.GetType() == documents.DocumentType_CODING_ASSISTANT {
			return d
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

func check(err error) {
	if err != nil {
		fmt.Println(out.Error(fmt.Sprintf("%v", err)))
		os.Exit(1)
	}
}
