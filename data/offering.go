package data

func FilterOfferings(vs *[]*Offering, f func(*Offering) bool) []*Offering {
	vsf := make([]*Offering, 0)
	for _, v := range *vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
