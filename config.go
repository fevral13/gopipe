package main

import (
	"strings"

	"github.com/spf13/viper"
)

// GP_URL=https://git....
// GP_API_KEY=glpat-....
// GP_PROJECT_ID=56
// GP_DELAY=10
// GP_MAIN_BRANCHES=development,release-23.12.x,release-23.11.x,release-23.10.x

type AppConfig struct {
	apiUrl       string
	apiKey       string
	mainBranches []string
	projectId    int
	delay        int
}

func getConfig() AppConfig {
	viper.SetEnvPrefix("gp")
	_ = viper.BindEnv("url")
	viper.SetDefault("url", "https://gitlab.com/")

	_ = viper.BindEnv("api_key")
	_ = viper.BindEnv("project_id")
	_ = viper.BindEnv("delay")
	viper.SetDefault("delay", 10)

	_ = viper.BindEnv("main_branches")
	mainBranchesString := viper.GetString("main_branches")
	mainBranches := strings.Split(mainBranchesString, ",")

	config := AppConfig{
		apiUrl:       viper.GetString("url"),
		apiKey:       viper.GetString("api_key"),
		projectId:    viper.GetInt("project_id"),
		delay:        viper.GetInt("delay"),
		mainBranches: mainBranches,
	}
	if config.apiKey == "" {
		panic("API key is not found in env. Export GP_API_KEY")
	}
	if config.projectId == 0 {
		panic("Project Id is not found in env. Export GP_PROJECT_ID")
	}

	return config
}
