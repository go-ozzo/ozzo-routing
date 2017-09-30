package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/golang/gddo/log"
)

const (
	gaeProjectEnvVar = "GCLOUD_PROJECT"
	gaAccountEnvVar  = "GA_ACCOUNT"
)

const (
	// Server Config
	ConfigProject           = "project"
	ConfigTrustProxyHeaders = "trust_proxy_headers"
	ConfigBindAddress       = "http"
	ConfigAssetsDir         = "assets"
	ConfigRobotThreshold    = "robot"
	ConfigGCELogName        = "gce_log_name"

	// Database Config
	ConfigDBServer      = "db-server"
	ConfigDBIdleTimeout = "db-idle-timeout"
	ConfigDBLog         = "db-log"
	ConfigGAERemoteAPI  = "remoteapi-endpoint"

	// Display Config
	ConfigSidebar        = "sidebar"
	ConfigSourcegraphURL = "sourcegraph_url"
	ConfigDefaultGOOS    = "default_goos"
	ConfigGAAccount      = "ga_account"

	// Crawl Config
	ConfigMaxAge          = "max_age"
	ConfigGetTimeout      = "get_timeout"
	ConfigFirstGetTimeout = "first_get_timeout"
	ConfigGithubInterval  = "github_interval"
	ConfigCrawlInterval   = "crawl_interval"
	ConfigDialTimeout     = "dial_timeout"
	ConfigRequestTimeout  = "request_timeout"
	ConfigMemcacheAddr    = "memcache_addr"
)

// Initialize configuration
func init() {
	ctx := context.Background()

	// Gather information from execution environment.
	if os.Getenv(gaeProjectEnvVar) != "" {
		viper.Set("on_appengine", true)
	} else {
		viper.Set("on_appengine", false)
	}
	if metadata.OnGCE() {
		gceProjectAttributeDefault(ctx, viper.GetViper(), ConfigGAAccount, "ga-account")
		gceProjectAttributeDefault(ctx, viper.GetViper(), ConfigGCELogName, "gce-log-name")
		if id, err := metadata.ProjectID(); err != nil {
			log.Warn(ctx, "failed to retrieve project ID", "error", err)
		} else {
			viper.SetDefault(ConfigProject, id)
		}
	}

	// Setup command line flags
	flags := buildFlags()
	flags.Parse(os.Args)
	if err := viper.BindPFlags(flags); err != nil {
		panic(err)
	}

	// Also fetch from enviorment
	viper.SetEnvPrefix("gddo")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	viper.BindEnv(ConfigProject, gaeProjectEnvVar)
	viper.BindEnv(ConfigGAAccount, gaAccountEnvVar)

	// Read from config.
	readViperConfig(ctx)

	// Set defaults based on other configs
	setDefaults()

	log.Info(ctx, "config values loaded", "values", viper.AllSettings())
}

func gceProjectAttributeDefault(ctx context.Context, v *viper.Viper, cfg, attr string) {
	val, err := metadata.ProjectAttributeValue(attr)
	if err != nil {
		if _, undef := err.(metadata.NotDefinedError); !undef {
			log.Warn(ctx, "failed to query metadata", "key", attr, "error", err)
		}
		return
	}
	v.SetDefault(cfg, val)
}

// setDefaults sets defaults for configuration options that depend on other
// configuration options. This allows for smart defaults but allows for
// overrides.
func setDefaults() {
	// ConfigGAERemoteAPI is based on project.
	project := viper.GetString(ConfigProject)
	if project != "" {
		defaultEndpoint := fmt.Sprintf("serviceproxy-dot-%s.appspot.com", project)
		viper.SetDefault(ConfigGAERemoteAPI, defaultEndpoint)
	}
}

func buildFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("default", pflag.ExitOnError)

	flags.StringP("config", "c", "", "path to motd config file")
	flags.String(ConfigProject, "", "Google Cloud Platform project used for Google services")
	// TODO(stephenmw): flags.Bool("enable-admin-pages", false, "When true, enables /admin pages")
	flags.Float64(ConfigRobotThreshold, 100, "Request counter threshold for robots.")
	flags.String(ConfigAssetsDir, filepath.Join(defaultBase("github.com/golang/gddo/gddo-server"), "assets"), "Base directory for templates and static files.")
	flags.Duration(ConfigGetTimeout, 8*time.Second, "Time to wait for package update from the VCS.")
	flags.Duration(ConfigFirstGetTimeout, 5*time.Second, "Time to wait for first fetch of package from the VCS.")
	flags.Duration(ConfigMaxAge, 24*time.Hour, "Update package documents older than this age.")
	flags.String(ConfigBindAddress, ":8080", "Listen for HTTP connections on this address.")
	flags.Bool(ConfigSidebar, false, "Enable package page sidebar.")
	flags.String(ConfigDefaultGOOS, "", "Default GOOS to use when building package documents.")
	flags.Bool(ConfigTrustProxyHeaders, false, "If enabled, identify the remote address of the request using X-Real-Ip in header.")
	flags.String(ConfigSourcegraphURL, "https://sourcegraph.com", "Link to global uses on Sourcegraph based at this URL (no need for trailing slash).")
	flags.Duration(ConfigGithubInterval, 0, "Github updates crawler sleeps for this duration between fetches. Zero disables the crawler.")
	flags.Duration(ConfigCrawlInterval, 0, "Package updater sleeps for this duration between package updates. Zero disables updates.")
	flags.Duration(ConfigDialTimeout, 5*time.Second, "Timeout for dialing an HTTP connection.")
	flags.Duration(ConfigRequestTimeout, 20*time.Second, "Time out for roundtripping an HTTP request.")
	flags.String(ConfigDBServer, "redis://127.0.0.1:6379", "URI of Redis server.")
	flags.Duration(ConfigDBIdleTimeout, 250*time.Second, "Close Redis connections after remaining idle for this duration.")
	flags.Bool(ConfigDBLog, false, "Log database commands")
	flags.String(ConfigMemcacheAddr, "", "Address in the format host:port gddo uses to point to the memcache backend.")
	flags.String(ConfigGAERemoteAPI, "", "Remoteapi endpoint for App Engine Search. Defaults to serviceproxy-dot-${project}.appspot.com.")

	return flags
}

// readViperConfig finds and then parses a config file. It will log.Fatal if the
// config file was specified or could not parse. Otherwise it will only warn
// that it failed to load the config.
func readViperConfig(ctx context.Context) {
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc")
	viper.SetConfigName("gddo")
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	}

	if err := viper.ReadInConfig(); err != nil {
		// If a config exists but could not be parsed, we should bail.
		if _, ok := err.(viper.ConfigParseError); ok {
			log.Fatal(ctx, "failed to parse config", "error", err)
		}

		// If the user specified a config file location in flags or env and
		// we failed to load it, we should bail. If not, it is just a warning.
		if viper.GetString("config") != "" {
			log.Fatal(ctx, "failed to load configuration file", "error", err)
		} else {
			log.Warn(ctx, "failed to load configuration file", "error", err)
		}
	} else {
		log.Info(ctx, "loaded configuration file successfully", "path", viper.ConfigFileUsed())
	}
}
