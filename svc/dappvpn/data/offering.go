package data

// OfferingParams is an offering additional parameters for VPN.
type OfferingParams struct {
	MinUploadMbits   float32 `json:"minUploadMbits"`
	MinDownloadMbits float32 `json:"minDownloadMbits"`
}
