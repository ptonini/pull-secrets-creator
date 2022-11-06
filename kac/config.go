/*
 * Kubernetes Admission Controller.
 * Copyright (C) 2022 Pedro Tonini
 * mailto:pedro DOT tonini AT hotmail DOT com
 *
 * Kubernetes Admission Controller is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 3 of the License, or (at your option) any later version.
 *
 * Kubernetes Admission Controller is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
 */

package kac

import (
	"encoding/json"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"strings"
)

type Config struct {
	ImagePullSecret map[string][]byte
}

func LoadConfig(configFile string) error {
	log.Printf("Loding config")
	return readConfig(configFile)
}

func readConfig(configFile string) error {
	viper.SetConfigName(strings.Split(filepath.Base(configFile), ".")[0])
	viper.AddConfigPath(filepath.Dir(configFile))
	viper.MustBindEnv("imagepullsecret", "IMAGE_PULL_SECRETS")
	return viper.ReadInConfig()
}

func getConfig() (*Config, error) {
	var config Config
	var err error
	switch viper.Get("imagepullsecret").(type) {
	case string:
		i := viper.GetString("imagepullsecret")
		err = json.Unmarshal([]byte(i), &config.ImagePullSecret)
	default:
		i, _ := json.Marshal(viper.GetStringMapString("imagepullsecret"))
		err = json.Unmarshal(i, &config.ImagePullSecret)
	}
	return &config, err
}
