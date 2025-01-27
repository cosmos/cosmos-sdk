package types

import "errors"

func (p *Permissions) Validation() error {
	switch {
	case p.Level == Permissions_LEVEL_SOME_MSGS:
		// if permission is some msg, LimitTypeUrls array must not be empty
		if len(p.LimitTypeUrls) == 0 {
			return errors.New("LimitTypeUrls of LEVEL_SOME_MSGS should NOT be empty")
		}

		p.LimitTypeUrls = MsgTypeURLValidation(p.LimitTypeUrls)
	case p.Level == Permissions_LEVEL_ALL_MSGS || p.Level == Permissions_LEVEL_SUPER_ADMIN:
		// if permission is all msg or super admin, LimitTypeUrls array clear
		// all p.LimitTypeUrls since we not use this field
		p.LimitTypeUrls = nil
	default:
	}

	return nil
}

func MsgTypeURLValidation(urls []string) []string {
	for idx, url := range urls {
		if len(url) == 0 {
			continue
		}
		if url[0] != '/' {
			urls[idx] = "/" + url
		}
	}
	return urls
}
