package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

type registryResolver struct {
	client *http.Client
}

var Registry = &registryResolver{
	client: &http.Client{
		// give it a maximum 5 seconds to resolve the module
		Timeout: time.Second * 5,
	},
}

type moduleVersions struct {
	Modules []struct {
		Versions []struct {
			Version string `json:"version"`
		} `json:"versions"`
	} `json:"modules"`
}

const registryHostname = "registry.terraform.io"

func (r *registryResolver) Resolve(ctx context.Context, target fs.FS, opt Options) (filesystem fs.FS, prefix string, downloadPath string, applies bool, err error) {

	if !opt.AllowDownloads {
		return
	}

	inputVersion := opt.Version
	source, relativePath, _ := strings.Cut(opt.Source, "//")
	parts := strings.Split(source, "/")
	if len(parts) < 3 || len(parts) > 4 {
		return
	}

	rootParts := parts
	if len(rootParts) > 3 {
		rootParts = rootParts[:3]
		relativePath = strings.Join(rootParts[3:], "/")
	}

	moduleName := strings.Join(rootParts, "/")

	if opt.Version != "" {
		versionUrl := fmt.Sprintf("https://%s/v1/modules/%s/versions", registryHostname, moduleName)
		opt.Debug("Requesting module versions from registry using '%s'...", versionUrl)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionUrl, nil)
		if err != nil {
			return nil, "", "", true, err
		}
		resp, err := r.client.Do(req)
		if err != nil {
			return nil, "", "", true, err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return nil, "", "", true, fmt.Errorf("unexpected status code for versions endpoint: %d", resp.StatusCode)
		}
		var availableVersions moduleVersions
		if err := json.NewDecoder(resp.Body).Decode(&availableVersions); err != nil {
			return nil, "", "", true, err
		}

		opt.Version, err = resolveVersion(inputVersion, availableVersions)
		if err != nil {
			return nil, "", "", true, err
		}
		opt.Debug("Found version '%s' for constraint '%s'", opt.Version, inputVersion)
	}

	var url string
	if opt.Version == "" {
		url = fmt.Sprintf("https://%s/v1/modules/%s/download", registryHostname, moduleName)
	} else {
		url = fmt.Sprintf("https://%s/v1/modules/%s/%s/download", registryHostname, moduleName, opt.Version)
	}

	opt.Debug("Requesting module source from registry using '%s'...", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", "", true, err
	}
	if opt.Version != "" {
		req.Header.Set("X-Terraform-Version", opt.Version)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, "", "", true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return nil, "", "", true, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	opt.Source = resp.Header.Get("X-Terraform-Get")
	opt.Debug("Module '%s' resolved via registry to new source: '%s'", opt.Name, opt.Source)
	opt.RelativePath = relativePath
	filesystem, prefix, downloadPath, _, err = Remote.Resolve(ctx, target, opt)
	if err != nil {
		return nil, "", "", true, err
	}

	return filesystem, prefix, downloadPath, true, nil
}

func resolveVersion(input string, versions moduleVersions) (string, error) {
	if len(versions.Modules) != 1 {
		return "", fmt.Errorf("1 module expected, found %d", len(versions.Modules))
	}
	if len(versions.Modules[0].Versions) == 0 {
		return "", fmt.Errorf("no available versions for module")
	}
	constraints, err := semver.NewConstraint(input)
	if err != nil {
		return "", err
	}
	var realVersions semver.Collection
	for _, rawVersion := range versions.Modules[0].Versions {
		realVersion, err := semver.NewVersion(rawVersion.Version)
		if err != nil {
			continue
		}
		realVersions = append(realVersions, realVersion)
	}
	sort.Sort(sort.Reverse(realVersions))
	for _, realVersion := range realVersions {
		if constraints.Check(realVersion) {
			return realVersion.String(), nil
		}
	}
	return "", fmt.Errorf("no available versions for module constraint '%s'", input)
}
