package kuper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type StoreInfo struct {
	StoreID      int
	StoreName    string
	StoreAddress string
	RetailerName string
}

type storeResp struct {
	Store struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Location struct {
			FullAddress string `json:"full_address"`
			City        string `json:"city"`
			Street      string `json:"street"`
			Building    string `json:"building"`
		} `json:"location"`
		Retailer struct {
			Name string `json:"name"`
		} `json:"retailer"`
	} `json:"store"`
}

func (s *service) GetStore(ctx context.Context, storeID int) (StoreInfo, error) {
	url := fmt.Sprintf("%s/api/stores/%d", s.baseURL, storeID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return StoreInfo{}, err
	}
	s.applyDefaultHeaders(req)

	resp, err := s.transport.Do(req)
	if err != nil {
		return StoreInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return StoreInfo{}, fmt.Errorf("GetStore: статус=%d body=%s", resp.StatusCode, string(b))
	}

	var out storeResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return StoreInfo{}, err
	}

	addr := out.Store.Location.FullAddress
	if addr == "" {
		addr = fmt.Sprintf("%s, %s %s", out.Store.Location.City, out.Store.Location.Street, out.Store.Location.Building)
	}

	name := out.Store.Name
	if name == "" {
		name = out.Store.FullName
	}

	return StoreInfo{
		StoreID:      out.Store.ID,
		StoreName:    name,
		StoreAddress: addr,
		RetailerName: out.Store.Retailer.Name,
	}, nil
}
