package project

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/cmd/project/focus"
	"github.com/devplaninc/devplan-cli/internal/components/selector"
	"github.com/devplaninc/devplan-cli/internal/devplan"
	"github.com/devplaninc/devplan-cli/internal/out"
	"github.com/devplaninc/devplan-cli/internal/utils/git"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/artifacts"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/grouping"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
)

const (
	nowSection   = "now-projects"
	nextSection  = "next-projects"
	laterSection = "later-projects"
)

var Cmd = create()

func init() {
	Cmd.AddCommand(focus.Cmd)
}

func mainGroupID(companyID int32) string {
	return fmt.Sprintf("%v-projects", companyID)
}

func create() *cobra.Command {
	return &cobra.Command{
		Use:   "project",
		Short: "List all available projects",
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
			prompts, err := getPrompts(feature.GetId(), project.GetDocs())
			check(err)
			fmt.Printf("Found %v prompts\n", len(prompts))
			sumResp, err := cl.GetRepoSummaries(company.GetId())
			check(err)
			summary := getMatchingSummary(repo, sumResp.GetSummaries())
			fmt.Printf("Matching summary found: %v\n", summary != nil)
		},
	}
}

func getMatchingSummary(repo git.RepoInfo, summaries []*artifacts.ArtifactRepoSummary) *artifacts.ArtifactRepoSummary {
	for _, s := range summaries {
		if repo.MatchesName(s.GetRepoName()) {
			return s
		}
	}
	return nil
}

func getPrompts(featureID string, docs []*documents.DocumentEntity) ([]*documents.DocumentEntity, error) {
	codeAssist := getCodingAssistant(docs)
	if codeAssist == nil {
		return nil, nil
	}
	var result []*documents.DocumentEntity
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
			result = append(result, d)
		}
	}
	return result, nil
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
