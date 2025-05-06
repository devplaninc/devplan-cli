package globals

import "github.com/spf13/viper"

var Domain string

const (
	APIkeyConfig      = "apikey"
	LastCompanyConfig = "lastCompany"
	LastProjectConfig = "lastProject"
)

// GetAPIKey gets the API key from the configuration
func GetAPIKey() (string, error) {
	return viper.GetString(APIkeyConfig), nil
}

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
