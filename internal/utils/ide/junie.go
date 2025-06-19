package ide

import (
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/documents"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/integrations"
)

func createJunieRules(featurePrompt *documents.DocumentEntity, repoSummary *integrations.RepositorySummary) error {
	return createMdRules(".junie", featurePrompt, repoSummary)
}
