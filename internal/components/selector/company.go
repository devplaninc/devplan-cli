package selector

import (
	"fmt"
	"github.com/devplaninc/webapp/golang/pb/api/devplan/types/company"
)

func Company(companies []*company.CompanyDetails, props Props, chosen int32) (*company.CompanyDetails, error) {
	if chosen >= 0 {
		for _, c := range companies {
			if c.GetId() == chosen {
				return c, nil
			}
		}
		return nil, fmt.Errorf("company [%v] not found among user companies", chosen)
	}
	if len(companies) == 1 {
		return companies[0], nil
	}
	var items []item
	for _, c := range companies {
		items = append(items, item{id: fmt.Sprintf("%v", c.GetId()), title: c.GetName()})
	}
	companyID, err := runSelector("company", items, props)
	if err != nil {
		return nil, err
	}
	if companyID == "" {
		return nil, nil
	}
	for _, c := range companies {
		if fmt.Sprintf("%v", c.GetId()) == companyID {
			return c, nil
		}
	}
	return nil, nil
}
