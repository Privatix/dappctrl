package data

func FilterSessions(vs *[]*Session, f func(*Session) bool) []*Session {
	vsf := make([]*Session, 0)
	for _, v := range *vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
