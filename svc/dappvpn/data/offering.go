package data

// OfferingParams is an offering additional parameters for VPN.
type OfferingParams struct {
	MinUploadMbits   int `json:"minUploadMbits"`
	MinDownloadMbits int `json:"minDownloadMbits"`
}
