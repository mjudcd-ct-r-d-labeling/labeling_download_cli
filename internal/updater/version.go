package updater

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var numericVersion = regexp.MustCompile(`^[0-9]+(\.[0-9]+){2,3}$`)

func compareVersions(a, b string) (int, error) {
	if !numericVersion.MatchString(a) || !numericVersion.MatchString(b) {
		return 0, fmt.Errorf("invalid version")
	}
	ap, bp := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(ap) || i < len(bp); i++ {
		var av, bv uint64
		var err error
		if i < len(ap) {
			av, err = strconv.ParseUint(ap[i], 10, 64)
			if err != nil {
				return 0, err
			}
		}
		if i < len(bp) {
			bv, err = strconv.ParseUint(bp[i], 10, 64)
			if err != nil {
				return 0, err
			}
		}
		if av > bv {
			return 1, nil
		}
		if av < bv {
			return -1, nil
		}
	}
	return 0, nil
}

func latestVersion(ctx context.Context, c *http.Client, address string) (string, error) {
	latest := ""
	for page := 1; page <= 100; page++ {
		var tags []struct {
			Name string `json:"name"`
		}
		if err := getJSON(ctx, c, fmt.Sprintf("%s?per_page=100&page=%d", address, page), &tags); err != nil {
			return "", err
		}
		for _, tag := range tags {
			if !numericVersion.MatchString(tag.Name) {
				continue
			}
			if _, err := compareVersions(tag.Name, tag.Name); err != nil {
				continue
			}
			if latest == "" {
				latest = tag.Name
				continue
			}
			if cmp, err := compareVersions(tag.Name, latest); err == nil && cmp > 0 {
				latest = tag.Name
			}
		}
		if len(tags) < 100 {
			if latest == "" {
				return "", fmt.Errorf("no published CLI version was found")
			}
			return latest, nil
		}
	}
	return "", fmt.Errorf("too many release tags; use the official installer")
}
