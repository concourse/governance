package governance

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Contributors map[string]Person
	Teams        map[string]Team
	Repos        map[string]Repo
}

type Person struct {
	Name    string            `yaml:"name"`
	GitHub  string            `yaml:"github"`
	Discord string            `yaml:"discord,omitempty"`
	Email   string            `yaml:"email,omitempty"`
	Repos   map[string]string `yaml:"repos,omitempty"`
}

type Team struct {
	Name             string   `yaml:"name"`
	Purpose          string   `yaml:"purpose"`
	Responsibilities []string `yaml:"responsibilities"`

	Discord Discord `yaml:"discord,omitempty"`

	AllContributors bool     `yaml:"all_contributors"`
	RawMembers      []string `yaml:"members"`

	RequiresEmail bool `yaml:"requires_email,omitempty"`

	RawRepoPermission string   `yaml:"repo_permission"`
	Repos             []string `yaml:"repos,omitempty"`
}

func (team Team) Members(cfg *Config) map[string]Person {
	if team.AllContributors {
		return cfg.Contributors
	} else {
		members := map[string]Person{}
		for _, m := range team.RawMembers {
			members[m] = cfg.Contributors[m]
		}

		return members
	}
}

func (team Team) RepoPermission() RepoPermission {
	if team.RawRepoPermission == "" {
		return RepoPermissionMaintain
	}

	return permission3to4(team.RawRepoPermission)
}

type Discord struct {
	Role     string `yaml:"role,omitempty"`
	Color    int    `yaml:"color,omitempty"`
	Priority int    `yaml:"priority,omitempty"`

	AddedPermissions DiscordPermissionSet `yaml:"added_permissions,omitempty"`

	// if set, the role will never be removed. this is primarily to
	// grandfather in users who have roles which predated the governance
	// automation, i.e. the 'contributors' role.
	Sticky bool `yaml:"sticky,omitempty"`
}

// 1. copied from https://discord.com/developers/docs/topics/permissions#permissions-bitwise-permission-flags
// 2. replaced GUILD with SERVER
var DiscordPermissions = map[string]int64{
	"CREATE_INSTANT_INVITE": 0x00000001,
	"KICK_MEMBERS":          0x00000002,
	"BAN_MEMBERS":           0x00000004,
	"ADMINISTRATOR":         0x00000008,
	"MANAGE_CHANNELS":       0x00000010,
	"MANAGE_SERVER":         0x00000020,
	"ADD_REACTIONS":         0x00000040,
	"VIEW_AUDIT_LOG":        0x00000080,
	"PRIORITY_SPEAKER":      0x00000100,
	"STREAM":                0x00000200,
	"VIEW_CHANNEL":          0x00000400,
	"SEND_MESSAGES":         0x00000800,
	"SEND_TTS_MESSAGES":     0x00001000,
	"MANAGE_MESSAGES":       0x00002000,
	"EMBED_LINKS":           0x00004000,
	"ATTACH_FILES":          0x00008000,
	"READ_MESSAGE_HISTORY":  0x00010000,
	"MENTION_EVERYONE":      0x00020000,
	"USE_EXTERNAL_EMOJIS":   0x00040000,
	"VIEW_SERVER_INSIGHTS":  0x00080000,
	"CONNECT":               0x00100000,
	"SPEAK":                 0x00200000,
	"MUTE_MEMBERS":          0x00400000,
	"DEAFEN_MEMBERS":        0x00800000,
	"MOVE_MEMBERS":          0x01000000,
	"USE_VAD":               0x02000000,
	"CHANGE_NICKNAME":       0x04000000,
	"MANAGE_NICKNAMES":      0x08000000,
	"MANAGE_ROLES":          0x10000000,
	"MANAGE_WEBHOOKS":       0x20000000,
	"MANAGE_EMOJIS":         0x40000000,
}

// defaults copied from newly created role; may be worth tuning later
type DiscordPermissionSet []string

func (set DiscordPermissionSet) Permissions() (int64, error) {
	var permissions int64
	for _, permission := range set {
		bits, found := DiscordPermissions[permission]
		if !found {
			return 0, fmt.Errorf("unknown permission: %s", permission)
		}

		permissions |= bits
	}

	return permissions, nil
}

var TeamRoleBasePermissions = DiscordPermissionSet{
	"VIEW_CHANNEL",
	"CREATE_INSTANT_INVITE",
	"CHANGE_NICKNAME",
	"SEND_MESSAGES",
	"EMBED_LINKS",
	"ATTACH_FILES",
	"ADD_REACTIONS",
	"USE_EXTERNAL_EMOJIS",
	"MENTION_EVERYONE",
	"READ_MESSAGE_HISTORY",
	"SEND_TTS_MESSAGES",
	"CONNECT",
	"SPEAK",
	"STREAM",
	"USE_VAD",
}

type Repo struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Private     bool     `yaml:"private,omitempty"`
	Topics      []string `yaml:"topics,omitempty"`
	HomepageURL string   `yaml:"homepage_url,omitempty"`
	HasIssues   bool     `yaml:"has_issues,omitempty"`
	HasProjects bool     `yaml:"has_projects,omitempty"`
	HasWiki     bool     `yaml:"has_wiki,omitempty"`

	Pages *RepoPages `yaml:"pages,omitempty"`

	BranchProtection []RepoBranchProtection `yaml:"branch_protection,omitempty"`

	Labels []RepoLabel `yaml:"labels,omitempty"`

	DeployKeys []RepoDeployKey `yaml:"deploy_keys"`
}

type RepoPages struct {
	CNAME  string `yaml:"cname,omitempty"`
	Branch string `yaml:"branch"`
	Path   string `yaml:"path,omitempty"`
}

type RepoBranchProtection struct {
	Pattern string `yaml:"pattern"`

	AllowsDeletions bool `yaml:"allows_deletions,omitempty"`

	RequiredChecks []string `yaml:"required_checks,omitempty"`
	StrictChecks   bool     `yaml:"strict_checks,omitempty"`

	RequiredReviews         int  `yaml:"required_reviews,omitempty"`
	DismissStaleReviews     bool `yaml:"dismiss_stale_reviews,omitempty"`
	RequireCodeOwnerReviews bool `yaml:"require_code_owner_reviews,omitempty"`
}

type RepoLabel struct {
	Name  string `yaml:"name"`
	Color int    `yaml:"color"`
}

type RepoDeployKey struct {
	Title     string `yaml:"title"`
	PublicKey string `yaml:"public_key"`
	Writable  bool   `yaml:"writable"`
}

func LoadConfig(tree fs.FS) (*Config, error) {
	personFiles, err := fs.ReadDir(tree, "contributors")
	if err != nil {
		return nil, err
	}

	contributors := map[string]Person{}
	for _, f := range personFiles {
		fn := filepath.Join("contributors", f.Name())

		file, err := tree.Open(fn)
		if err != nil {
			return nil, err
		}

		var person Person
		err = decode(file, &person)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", fn, err)
		}

		contributors[strings.TrimSuffix(f.Name(), ".yml")] = person
	}

	teamFiles, err := fs.ReadDir(tree, "teams")
	if err != nil {
		return nil, err
	}

	teams := map[string]Team{}
	for _, f := range teamFiles {
		fn := filepath.Join("teams", f.Name())

		file, err := tree.Open(fn)
		if err != nil {
			return nil, err
		}

		var team Team
		err = decode(file, &team)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", fn, err)
		}

		teams[strings.TrimSuffix(f.Name(), ".yml")] = team
	}

	repoFiles, err := fs.ReadDir(tree, "repos")
	if err != nil {
		return nil, err
	}

	repos := map[string]Repo{}
	for _, f := range repoFiles {
		fn := filepath.Join("repos", f.Name())

		file, err := tree.Open(fn)
		if err != nil {
			return nil, err
		}

		var repo Repo
		err = decode(file, &repo)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", fn, err)
		}

		repos[strings.TrimSuffix(f.Name(), ".yml")] = repo
	}

	return &Config{
		Contributors: contributors,
		Teams:        teams,
		Repos:        repos,
	}, nil
}

// collapse word-wrapped string YAML blocks
func sanitize(str string) string {
	return strings.TrimSpace(strings.Join(strings.Split(str, "\n"), " "))
}

func permission3to4(v3permission string) RepoPermission {
	switch v3permission {
	case "pull":
		return RepoPermissionRead
	case "push":
		return RepoPermissionWrite
	case "admin":
		return RepoPermissionAdmin
	case "maintain":
		return RepoPermissionMaintain
	case "triage":
		return RepoPermissionTriage
	default:
		return "INVALID"
	}
}

func permission4to3(v4permission RepoPermission) string {
	switch v4permission {
	case RepoPermissionRead:
		return "pull"
	case RepoPermissionWrite:
		return "push"
	case RepoPermissionAdmin:
		return "admin"
	case RepoPermissionMaintain:
		return "maintain"
	case RepoPermissionTriage:
		return "triage"
	default:
		return "invalid"
	}
}

func decode(stream io.Reader, dest interface{}) error {
	dec := yaml.NewDecoder(stream)
	dec.SetStrict(true)
	return dec.Decode(dest)
}
