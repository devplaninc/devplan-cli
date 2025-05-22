package selector

import (
	"fmt"
	"github.com/devplaninc/devplan-cli/internal/utils/prefs"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/company"
)

func Company(companies []*company.CompanyDetails, props Props, chosen int32) (*company.CompanyDetails, error) {
	// If a company ID is explicitly provided, use it
	if chosen >= 0 {
		for _, c := range companies {
			if c.GetId() == chosen {
				// Save the selected company ID to preferences
				prefs.SetLastCompanyID(c.GetId())
				return c, nil
			}
		}
		return nil, fmt.Errorf("company [%v] not found among user companies", chosen)
	}

	// If there's only one company, use it
	if len(companies) == 1 {
		// Save the selected company ID to preferences
		prefs.SetLastCompanyID(companies[0].GetId())
		return companies[0], nil
	}

	// Try to get the last selected company from preferences
	lastCompanyID := prefs.GetLastCompanyID()
	if lastCompanyID > 0 {
		// Validate that the company still exists and is accessible
		for _, c := range companies {
			if c.GetId() == lastCompanyID {
				// Use the last selected company as default
				chosen = lastCompanyID
				break
			}
		}
	}

	var items []item
	var defaultIndex int
	for i, c := range companies {
		items = append(items, item{id: fmt.Sprintf("%v", c.GetId()), title: c.GetName()})
		if c.GetId() == chosen {
			defaultIndex = i
		}
	}

	companyID, err := runSelector("company", items, props, defaultIndex)
	if err != nil {
		return nil, err
	}
	if companyID == "" {
		return nil, nil
	}

	for _, c := range companies {
		if fmt.Sprintf("%v", c.GetId()) == companyID {
			// Save the selected company ID to preferences
			prefs.SetLastCompanyID(c.GetId())
			return c, nil
		}
	}
	return nil, nil
}
