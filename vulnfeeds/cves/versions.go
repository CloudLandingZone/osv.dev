// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cves

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/knqyf263/go-cpe/naming"
	"golang.org/x/exp/slices"
)

type GitCommit struct {
	Repo   string
	Commit string
}

type AffectedVersion struct {
	Introduced   string
	Fixed        string
	LastAffected string
}

type VersionInfo struct {
	IntroducedCommits   []GitCommit
	FixCommits          []GitCommit
	LimitCommits        []GitCommit
	LastAffectedCommits []GitCommit
	AffectedVersions    []AffectedVersion
}

type CPE struct {
	CPEVersion string
	Part       string
	Vendor     string
	Product    string
	Version    string
	Update     string
	Edition    string
	Language   string
	SWEdition  string
	TargetSW   string
	TargetHW   string
	Other      string
}

var (
	// TODO(apollock): read this from an external file
	InvalidRepos = []string{
		"https://github.com/abhiunix/goo-blog-App-CVE",
		"https://github.com/Accenture/AARO-Bugs",
		"https://github.com/active-labs/Advisories",
		"https://github.com/afeng2016-s/CVE-Request",
		"https://github.com/agadient/SERVEEZ-CVE",
		"https://github.com/AlwaysHereFight/YZMCMSxss",
		"https://github.com/alwentiu/COVIDSafe-CVE-2020-12856",
		"https://github.com/ArianeBlow/Axelor_Stored_XSS",
		"https://github.com/beicheng-maker/vulns",
		"https://github.com/BigTiger2020/74CMS",
		"https://github.com/BlackFan/client-side-prototype-pollution",
		"https://github.com/blindkey/cve_like",
		"https://github.com/ciph0x01/Simple-Exam-Reviewer-Management-System-CVE",
		"https://github.com/cloudflare/advisories",
		"https://github.com/CVEProject/cvelist", // Heavily in Advisory URLs, sometimes shows up elsewhere
		"https://github.com/cve-vul/vul",
		"https://github.com/daaaalllii/cve-s",
		"https://github.com/DayiliWaseem/CVE-2022-39196-",
		"https://github.com/eddietcc/CVEnotes",
		"https://github.com/enesozeser/Vulnerabilities",
		"https://github.com/Fadavvi/CVE-2018-17431-PoC",
		"https://github.com/FCncdn/Appsmith-Js-Injection-POC",
		"https://github.com/fireeye/Vulnerability-Disclosures",
		"https://github.com/GitHubAssessments/CVE_Assessments_11_2019",
		"https://github.com/github/cvelist",        // Fork of https://github.com/CVEProject/cvelist
		"https://github.com/gitlabhq/gitlabhq",     // GitHub mirror, not canonical
		"https://github.com/google/oss-fuzz-vulns", // 8^)
		"https://github.com/Gr4y21/My-CVE-IDs",
		"https://github.com/hashicorp/terraform-enterprise-release-notes",
		"https://github.com/hemantsolo/CVE-Reference",
		"https://github.com/huclilu/CVE_Add",
		"https://github.com/i3umi3iei3ii/CentOS-Control-Web-Panel-CVE",
		"https://github.com/ianxtianxt/gitbook-xss",
		"https://github.com/itodaro/doorGets_cve",
		"https://github.com/jvz/test-cvelist",
		"https://github.com/Kenun99/CVE-batdappboomx",
		"https://github.com/kyrie403/Vuln",
		"https://github.com/lukaszstu/SmartAsset-CORS-CVE-2020-26527",
		"https://github.com/MacherCS/CVE_Evoh_Contract",
		"https://github.com/mandiant/Vulnerability-Disclosures",
		"https://github.com/martinkubecka/CVE-References",
		"https://github.com/mclab-hbrs/BBB-POC",
		"https://github.com/metaredteam/external-disclosures",
		"https://github.com/MrR3boot/CVE-Hunting",
		"https://github.com/N1ce759/74cmsSE-Arbitrary-File-Reading",
		"https://github.com/nepenthe0320/cve_poc",
		"https://github.com/Netflix/security-bulletins",
		"https://github.com/nextcloud/security-advisories",
		"https://github.com/nikip72/CVE-2021-39273-CVE-2021-39274",
		"https://github.com/nu11secur1ty/CVE-nu11secur1ty",
		"https://github.com/orangecertcc/security-research",
		"https://github.com/Orange-Cyberdefense/CVE-repository",
		"https://github.com/passtheticket/vulnerability-research",
		"https://github.com/post-cyberlabs/CVE-Advisory",
		"https://github.com/rapid7/metasploit-framework",
		"https://github.com/refi64/CVE-2020-25265-25266",
		"https://github.com/riteshgohil/My_CVE_References",
		"https://github.com/roughb8722/CVE-2021-3122-Details",
		"https://github.com/Ryan0lb/EC-cloud-e-commerce-system-CVE-application",
		"https://github.com/SaumyajeetDas/POC-of-CVE-2022-36271",
		"https://github.com/seb1055/cve-2020-27358-27359",
		"https://github.com/Security-AVS/-CVE-2021-26904",
		"https://github.com/seqred-s-a/gxdlmsdirector-cve",
		"https://github.com/sickcodes/security",
		"https://github.com/Snakinya/Vuln",
		"https://github.com/snyk/zip-slip-vulnerability",
		"https://github.com/soheilsamanabadi/vulnerability",
		"https://github.com/soheilsamanabadi/vulnerabilitys",
		"https://github.com/theyiyibest/Reflected-XSS-on-SockJS",
		"https://github.com/vQAQv/Request-CVE-ID-PoC",
		"https://github.com/vulnerabilities-cve/vulnerabilities",
		"https://github.com/wind-cyber/LJCMS-UserTraversal-Vulnerability",
		"https://github.com/wsummerhill/BSA-Radar_CVE-Vulnerabilities",
		"https://github.com/xiahao90/CVEproject",
		"https://github.com/xxhzz1/74cmsSE-Arbitrary-file-upload-vulnerability",
		"https://github.com/ycdxsb/Vuln",
		"https://github.com/YLoiK/74cmsSE-Arbitrary-file-upload-vulnerability",
		"https://github.com/z00z00z00/Safenet_SAC_CVE-2021-42056",
		"https://github.com/zer0yu/CVE_Request",
		"https://github.com/Zeyad-Azima/Issabel-stored-XSS",
		"https://gitlab.com/gitlab-org/gitlab-ce",      // redirects to gitlab-foss
		"https://gitlab.com/gitlab-org/gitlab-ee",      // redirects to gitlab
		"https://gitlab.com/gitlab-org/gitlab-foss",    // not the canonical source
		"https://gitlab.com/gitlab-org/omnibus-gitlab", // not the source
		"https://gitlab.com/gitlab-org/release",        // not the source
	}
	InvalidRepoRegex = `(?i)/(?:(?:CVEs?)|(?:CVE-\d{4}-\d{4,})|GitHubAssessments/.*)$`
)

// Returns the base repository URL for supported repository hosts.
func Repo(u string) (string, error) {
	var supportedHosts = []string{
		"github.com",
		"gitlab.org",
		"bitbucket.org",
	}
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	// Disregard the repos we know we don't like (by regex).
	matched, _ := regexp.MatchString(InvalidRepoRegex, u)
	if matched {
		return "", fmt.Errorf("%q matched invalid repo regexp", u)
	}

	for _, dr := range InvalidRepos {
		if strings.HasPrefix(u, dr) {
			return "", fmt.Errorf("%q found in denylist", u)
		}
	}

	// Were we handed a base repository URL from the get go?
	if slices.Contains(supportedHosts, parsedURL.Hostname()) {
		if len(strings.Split(strings.TrimSuffix(parsedURL.Path, "/"), "/")) == 3 {
			return fmt.Sprintf("%s://%s%s", parsedURL.Scheme,
					parsedURL.Hostname(),
					strings.TrimSuffix(parsedURL.Path, "/")),
				nil
		}
	}

	// cGit URLs are structured another way, e.g.
	// https://git.dpkg.org/cgit/dpkg/dpkg.git/commit/?id=faa4c92debe45412bfcf8a44f26e827800bb24be
	// https://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/commit/?id=817b8b9c5396d2b2d92311b46719aad5d3339dbe
	if strings.HasPrefix(parsedURL.Path, "/cgit") &&
		strings.HasSuffix(parsedURL.Path, "commit/") &&
		strings.HasPrefix(parsedURL.RawQuery, "id=") {
		repo := strings.TrimSuffix(parsedURL.Path, "/commit/")
		return fmt.Sprintf("%s://%s%s", parsedURL.Scheme,
			parsedURL.Hostname(), repo), nil
	}

	// GitWeb CGI URLs are structured very differently, e.g.
	// https://git.gnupg.org/cgi-bin/gitweb.cgi?p=libksba.git;a=commit;h=f61a5ea4e0f6a80fd4b28ef0174bee77793cf070 is another variation seen in the wild
	if strings.HasPrefix(parsedURL.Path, "/cgi-bin/gitweb.cgi") &&
		strings.HasPrefix(parsedURL.RawQuery, "p=") {
		params := strings.Split(parsedURL.RawQuery, ";")
		for _, param := range params {
			if !strings.HasPrefix(param, "p=") {
				continue
			}
			repo := strings.Split(param, "=")[1]
			return fmt.Sprintf("%s://%s/%s", parsedURL.Scheme, parsedURL.Hostname(), repo), nil
		}
	}

	// cgit.freedesktop.org is a special snowflake with enough repos to warrant special handling
	// it is a mirror of gitlab.freedesktop.org
	// https://cgit.freedesktop.org/xorg/lib/libXRes/commit/?id=c05c6d918b0e2011d4bfa370c321482e34630b17
	// https://cgit.freedesktop.org/xorg/lib/libXRes
	// http://cgit.freedesktop.org/spice/spice/refs/tags
	if parsedURL.Hostname() == "cgit.freedesktop.org" {
		if strings.HasSuffix(parsedURL.Path, "commit/") &&
			strings.HasPrefix(parsedURL.RawQuery, "id=") {
			repo := strings.TrimSuffix(parsedURL.Path, "/commit/")
			return fmt.Sprintf("https://gitlab.freedesktop.org%s",
				repo), nil
		}
		if strings.HasSuffix(parsedURL.Path, "refs/tags") {
			repo := strings.TrimSuffix(parsedURL.Path, "/refs/tags")
			return fmt.Sprintf("https://gitlab.freedesktop.org%s",
				repo), nil
		}
		if len(strings.Split(parsedURL.Path, "/")) == 4 {
			return fmt.Sprintf("https://gitlab.freedesktop.org%s",
				parsedURL.Path), nil
		}
	}

	// GitHub and GitLab commit and blob URLs are structured one way, e.g.
	// https://github.com/MariaDB/server/commit/b1351c15946349f9daa7e5297fb2ac6f3139e4a8
	// https://github.com/tensorflow/tensorflow/blob/master/tensorflow/core/ops/math_ops.cc
	// https://gitlab.freedesktop.org/virgl/virglrenderer/-/commit/b05bb61f454eeb8a85164c8a31510aeb9d79129c
	// https://gitlab.com/qemu-project/qemu/-/commit/4367a20cc4
	// https://gitlab.com/gitlab-org/cves/-/blob/master/2022/CVE-2022-2501.json
	//
	// This also supports GitHub tag URLs, e.g.
	// https://github.com/JonMagon/KDiskMark/releases/tag/3.1.0
	//
	// This also supports GitHub and Gitlab issue URLs, e.g.:
	// https://github.com/axiomatic-systems/Bento4/issues/755
	// https://gitlab.com/wireshark/wireshark/-/issues/18307
	//
	// This also supports GitHub Security Advisory URLs, e.g.
	// https://github.com/ballcat-projects/ballcat-codegen/security/advisories/GHSA-fv3m-xhqw-9m79
	if (parsedURL.Hostname() == "github.com" || strings.HasPrefix(parsedURL.Hostname(), "gitlab.")) &&
		(strings.Contains(parsedURL.Path, "commit") ||
			strings.Contains(parsedURL.Path, "blob") ||
			strings.Contains(parsedURL.Path, "releases/tag") ||
			strings.Contains(parsedURL.Path, "releases") ||
			strings.Contains(parsedURL.Path, "tags") ||
			strings.Contains(parsedURL.Path, "security/advisories") ||
			strings.Contains(parsedURL.Path, "issues")) {
		return fmt.Sprintf("%s://%s%s", parsedURL.Scheme,
				parsedURL.Hostname(),
				strings.Join(strings.Split(parsedURL.Path, "/")[0:3], "/")),
			nil
	}

	// GitHub pull request and comparison URLs are structured differently, e.g.
	// https://github.com/kovidgoyal/kitty/compare/v0.26.1...v0.26.2
	// https://gitlab.com/mayan-edms/mayan-edms/-/compare/development...master
	// https://git.drupalcode.org/project/views/-/compare/7.x-3.21...7.x-3.x
	if strings.Contains(parsedURL.Path, "compare") {
		return fmt.Sprintf("%s://%s%s", parsedURL.Scheme,
				parsedURL.Hostname(),
				strings.Join(strings.Split(parsedURL.Path, "/")[0:3], "/")),
			nil
	}

	// GitHub pull request URLs are structured differently, e.g.
	// https://github.com/google/osv.dev/pull/738
	if parsedURL.Hostname() == "github.com" &&
		strings.Contains(parsedURL.Path, "pull") {
		return fmt.Sprintf("%s://%s%s", parsedURL.Scheme,
				parsedURL.Hostname(),
				strings.Join(strings.Split(parsedURL.Path, "/")[0:3], "/")),
			nil
	}

	// Gitlab merge request URLs are structured differently, e.g.
	// https://gitlab.com/libtiff/libtiff/-/merge_requests/378
	if strings.HasPrefix(parsedURL.Hostname(), "gitlab.") &&
		strings.Contains(parsedURL.Path, "merge_requests") {
		return fmt.Sprintf("%s://%s%s", parsedURL.Scheme,
				parsedURL.Hostname(),
				strings.Join(strings.Split(parsedURL.Path, "/")[0:3], "/")),
			nil
	}

	// Bitbucket.org URLs are another snowflake, e.g.
	// https://bitbucket.org/ianb/pastescript/changeset/a19e462769b4
	// https://bitbucket.org/jespern/django-piston/commits/91bdaec89543/
	// https://bitbucket.org/openpyxl/openpyxl/commits/3b4905f428e1
	// https://bitbucket.org/snakeyaml/snakeyaml/pull-requests/35
	// https://bitbucket.org/snakeyaml/snakeyaml/issues/566
	// https://bitbucket.org/snakeyaml/snakeyaml/downloads/?tab=tags
	if parsedURL.Hostname() == "bitbucket.org" &&
		(strings.Contains(parsedURL.Path, "changeset") ||
			strings.Contains(parsedURL.Path, "downloads") ||
			strings.Contains(parsedURL.Path, "wiki") ||
			strings.Contains(parsedURL.Path, "issues") ||
			strings.Contains(parsedURL.Path, "security") ||
			strings.Contains(parsedURL.Path, "pull-requests") ||
			strings.Contains(parsedURL.Path, "commits")) {
		return fmt.Sprintf("%s://%s%s", parsedURL.Scheme,
				parsedURL.Hostname(),
				strings.Join(strings.Split(parsedURL.Path, "/")[0:3], "/")),
			nil
	}

	// If we get to here, we've encountered an unsupported URL.
	return "", fmt.Errorf("Repo(): unsupported URL: %s", u)
}

// Returns the commit ID from supported links.
func Commit(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	// cGit URLs are structured another way, e.g.
	// https://git.dpkg.org/cgit/dpkg/dpkg.git/commit/?id=faa4c92debe45412bfcf8a44f26e827800bb24be
	// https://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/commit/?id=817b8b9c5396d2b2d92311b46719aad5d3339dbe
	if strings.HasPrefix(parsedURL.Path, "/cgit") &&
		strings.HasSuffix(parsedURL.Path, "commit/") &&
		strings.HasPrefix(parsedURL.RawQuery, "id=") {
		return strings.Split(parsedURL.RawQuery, "=")[1], nil
	}

	// GitWeb cgi-bin URLs are structured another way, e.g.
	// https://git.gnupg.org/cgi-bin/gitweb.cgi?p=libksba.git;a=commit;h=f61a5ea4e0f6a80fd4b28ef0174bee77793cf070
	if strings.HasPrefix(parsedURL.Path, "/cgi-bin/gitweb.cgi") &&
		strings.Contains(parsedURL.RawQuery, "a=commit") {
		params := strings.Split(parsedURL.RawQuery, ";")
		for _, param := range params {
			if !strings.HasPrefix(param, "h=") {
				continue
			}
			return strings.Split(param, "=")[1], nil
		}
	}

	// GitHub and GitLab commit URLs are structured one way, e.g.
	// https://github.com/MariaDB/server/commit/b1351c15946349f9daa7e5297fb2ac6f3139e4a8
	// https://gitlab.freedesktop.org/virgl/virglrenderer/-/commit/b05bb61f454eeb8a85164c8a31510aeb9d79129c
	// https://gitlab.com/qemu-project/qemu/-/commit/4367a20cc4
	// and Bitbucket.org commit URLs are similiar yet slightly different:
	// https://bitbucket.org/openpyxl/openpyxl/commits/3b4905f428e1
	//
	// Some bitbucket.org commit URLs have been observed in the wild with a trailing /, which will
	// change the behaviour of path.Split(), so normalize the path to be tolerant of this.
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")
	directory, possibleCommitHash := path.Split(parsedURL.Path)
	if strings.HasSuffix(directory, "commit/") || strings.HasSuffix(directory, "commits/") {
		return possibleCommitHash, nil
	}

	// TODO(apollock): add support for resolving a GitHub PR to a commit hash

	// If we get to here, we've encountered an unsupported URL.
	return "", fmt.Errorf("Commit(): unsupported URL: %s", u)
}

// For URLs referencing commits in supported Git repository hosts, return a GitCommit.
func extractGitCommit(link string) *GitCommit {
	r, err := Repo(link)
	if err != nil {
		return nil
	}

	c, err := Commit(link)
	if err != nil {
		return nil
	}

	return &GitCommit{
		Repo:   r,
		Commit: c,
	}
}

func hasVersion(validVersions []string, version string) bool {
	if validVersions == nil || len(validVersions) == 0 {
		return true
	}
	return versionIndex(validVersions, version) != -1
}

func versionIndex(validVersions []string, version string) int {
	for i, cur := range validVersions {
		if cur == version {
			return i
		}
	}
	return -1
}

func nextVersion(validVersions []string, version string) (string, error) {
	idx := versionIndex(validVersions, version)
	if idx == -1 {
		return "", fmt.Errorf("Warning: %s is not a valid version", version)
	}

	idx += 1
	if idx >= len(validVersions) {
		return "", fmt.Errorf("Warning: %s does not have a version that comes after.", version)
	}

	return validVersions[idx], nil
}

func processExtractedVersion(version string) string {
	version = strings.Trim(version, ".")
	// Version should contain at least a "." or a number.
	if !strings.ContainsAny(version, ".") && !strings.ContainsAny(version, "0123456789") {
		return ""
	}

	return version
}

func extractVersionsFromDescription(validVersions []string, description string) ([]AffectedVersion, []string) {
	// Match:
	//  - x.x.x before x.x.x
	//  - x.x.x through x.x.x
	//  - through x.x.x
	//  - before x.x.x
	pattern := regexp.MustCompile(`(?i)([\w.+\-]+)?\s+(through|before)\s+(?:version\s+)?([\w.+\-]+)`)
	matches := pattern.FindAllStringSubmatch(description, -1)
	if matches == nil {
		return nil, []string{"Failed to parse versions from description"}
	}

	var notes []string
	var versions []AffectedVersion
	for _, match := range matches {
		// Trim periods that are part of sentences.
		introduced := processExtractedVersion(match[1])
		fixed := processExtractedVersion(match[3])
		if match[2] == "through" {
			// "Through" implies inclusive range, so the fixed version is the one that comes after.
			var err error
			fixed, err = nextVersion(validVersions, fixed)
			if err != nil {
				notes = append(notes, err.Error())
			}
		}

		if introduced == "" && fixed == "" {
			notes = append(notes, "Failed to match version range from description")
			continue
		}

		if introduced != "" && !hasVersion(validVersions, introduced) {
			notes = append(notes, fmt.Sprintf("Extracted version %s is not a valid version", introduced))
		}
		if fixed != "" && !hasVersion(validVersions, fixed) {
			notes = append(notes, fmt.Sprintf("Extracted version %s is not a valid version", fixed))
		}

		versions = append(versions, AffectedVersion{
			Introduced: introduced,
			Fixed:      fixed,
		})
	}

	return versions, notes
}

func cleanVersion(version string) string {
	// Versions can end in ":" for some reason.
	return strings.TrimRight(version, ":")
}

func ExtractVersionInfo(cve CVEItem, validVersions []string) (v VersionInfo, notes []string) {
	for _, reference := range cve.CVE.References.ReferenceData {
		if commit := extractGitCommit(reference.URL); commit != nil {
			v.FixCommits = append(v.FixCommits, *commit)
		}
	}

	gotVersions := false
	for _, node := range cve.Configurations.Nodes {
		if node.Operator != "OR" {
			continue
		}

		for _, match := range node.CPEMatch {
			if !match.Vulnerable {
				continue
			}

			introduced := ""
			fixed := ""
			lastaffected := ""
			if match.VersionStartIncluding != "" {
				introduced = cleanVersion(match.VersionStartIncluding)
			} else if match.VersionStartExcluding != "" {
				var err error
				introduced, err = nextVersion(validVersions, cleanVersion(match.VersionStartExcluding))
				if err != nil {
					notes = append(notes, err.Error())
				}
			}

			if match.VersionEndExcluding != "" {
				fixed = cleanVersion(match.VersionEndExcluding)
			} else if match.VersionEndIncluding != "" {
				var err error
				// Infer the fixed version from the next version after.
				fixed, err = nextVersion(validVersions, cleanVersion(match.VersionEndIncluding))
				if err != nil {
					notes = append(notes, err.Error())
					// if that inference failed, we know this version was definitely still vulnerable.
					lastaffected = cleanVersion(match.VersionEndIncluding)
					notes = append(notes, fmt.Sprintf("Using %s as last_affected version instead", cleanVersion(match.VersionEndIncluding)))
				}
			}

			if introduced == "" && fixed == "" {
				continue
			}

			if introduced != "" && !hasVersion(validVersions, introduced) {
				notes = append(notes, fmt.Sprintf("Warning: %s is not a valid introduced version", introduced))
			}

			if fixed != "" && !hasVersion(validVersions, fixed) {
				notes = append(notes, fmt.Sprintf("Warning: %s is not a valid fixed version", fixed))
			}

			gotVersions = true
			possibleNewAffectedVersion := AffectedVersion{
				Introduced:   introduced,
				Fixed:        fixed,
				LastAffected: lastaffected,
			}
			if slices.Contains(v.AffectedVersions, possibleNewAffectedVersion) {
				// Avoid appending duplicates
				continue
			}
			v.AffectedVersions = append(v.AffectedVersions, possibleNewAffectedVersion)
		}
	}
	if !gotVersions {
		var extractNotes []string
		v.AffectedVersions, extractNotes = extractVersionsFromDescription(validVersions, EnglishDescription(cve.CVE))
		notes = append(notes, extractNotes...)
		if len(v.AffectedVersions) > 0 {
			log.Printf("[%s] Extracted versions from description = %+v", cve.CVE.CVEDataMeta.ID, v.AffectedVersions)
		}
	}

	if len(v.AffectedVersions) == 0 {
		notes = append(notes, "No versions detected.")
	}

	if len(notes) != 0 && len(validVersions) > 0 {
		notes = append(notes, "Valid versions:")
		for _, version := range validVersions {
			notes = append(notes, "  - "+version)
		}
	}
	return v, notes
}

func CPEs(cve CVEItem) []string {
	var cpes []string
	for _, node := range cve.Configurations.Nodes {
		for _, match := range node.CPEMatch {
			cpes = append(cpes, match.CPE23URI)
		}
	}

	return cpes
}

// There are some weird and wonderful rules about quoting with strings in CPEs
// See 5.3.2 of NISTIR 7695 for more details
// https://nvlpubs.nist.gov/nistpubs/Legacy/IR/nistir7695.pdf
func RemoveQuoting(s string) (result string) {
	return strings.Replace(s, "\\", "", -1)
}

// Parse a well-formed CPE string into a struct.
func ParseCPE(formattedString string) (*CPE, error) {
	if !strings.HasPrefix(formattedString, "cpe:") {
		return nil, fmt.Errorf("%q does not have expected 'cpe:' prefix", formattedString)
	}

	wfn, err := naming.UnbindFS(formattedString)

	if err != nil {
		return nil, err
	}

	return &CPE{
		CPEVersion: strings.Split(formattedString, ":")[1],
		Part:       wfn.GetString("part"),
		Vendor:     RemoveQuoting(wfn.GetString("vendor")),
		Product:    RemoveQuoting(wfn.GetString("product")),
		Version:    RemoveQuoting(wfn.GetString("version")),
		Update:     wfn.GetString("update"),
		Edition:    wfn.GetString("edition"),
		Language:   wfn.GetString("language"),
		SWEdition:  wfn.GetString("sw_edition"),
		TargetSW:   wfn.GetString("target_sw"),
		TargetHW:   wfn.GetString("target_hw"),
		Other:      wfn.GetString("other")}, nil
}

// Normalize version strings found in CVE CPE Match data or Git tags.
// Use the same logic and behaviour as normalize_tag() osv/bug.py for consistency.
func NormalizeVersion(version string) (normalizedVersion string, e error) {
	// Keep in sync with the intent of https://github.com/google/osv.dev/blob/26050deb42785bc5a4dc7d802eac8e7f95135509/osv/bug.py#L31
	var validVersion = regexp.MustCompile(`(?i)(\d+|(?:rc|alpha|beta|preview)\d*)`)
	var validVersionText = regexp.MustCompile(`(?i)(?:rc|alpha|beta|preview)\d*`)
	components := validVersion.FindAllString(version, -1)
	if components == nil {
		return "", fmt.Errorf("%q is not a supported version", version)
	}
	// If the very first component happens to accidentally match the strings we support, remove it.
	// This is necessary because of the lack of negative lookbehind assertion support in RE2.
	if validVersionText.MatchString(components[0]) {
		components = slices.Delete(components, 0, 1)
	}
	normalizedVersion = strings.Join(components, "-")
	return normalizedVersion, e
}
