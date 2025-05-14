package globals

import (
	"github.com/devplaninc/devplan-cli/internal/utils/ide"
	"github.com/spf13/viper"
)

var Domain string

const (
	APIkeyConfig        = "apikey"
	LastCompanyConfig   = "lastCompany"
	LastProjectConfig   = "lastProject"
	LastAssistantConfig = "lastAssistant"
)

func GetLastCompany() int32 {
	return viper.GetInt32(LastCompanyConfig)
}

func SetLastCompany(companyID int32) {
	viper.Set(LastCompanyConfig, companyID)
}

func GetLastProject() string {
	return viper.GetString(LastProjectConfig)
}

func SetLastProject(projectID string) {
	viper.Set(LastProjectConfig, projectID)
}

func GetLastAssistant() (ide.Assistant, bool) {
	v := viper.GetString(LastAssistantConfig)
	return ide.Assistant(v), v != ""
}

func SetLastAssistant(asst ide.Assistant) {
	viper.Set(LastAssistantConfig, asst)
}
