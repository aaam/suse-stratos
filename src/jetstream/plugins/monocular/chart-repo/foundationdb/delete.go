/*
Copyright (c) 2019 The Helm Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package foundationdb

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Delete Deletes a chart repository from FoundationDB Document-Layer
func Delete(cmd *cobra.Command, args []string) {

	fdbURL, err := cmd.Flags().GetString("doclayer-url")
	if err != nil {
		log.Fatal(err)
	}
	fDB, err := cmd.Flags().GetString("doclayer-database")
	if err != nil {
		log.Fatal(err)
	}
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	log.Debugf("Creating client for FDB: %v, %v, %v", fdbURL, fDB, debug)
	clientOptions := options.Client().ApplyURI(fdbURL).SetMinPoolSize(10).SetMaxPoolSize(100)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := NewDocLayerClient(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Can't create client for FoundationDB document layer: %v URL provided was: %v", err, fdbURL)
		return
	}

	log.Debugf("Client created.")

	//Close db client after serving request
	var clientKeepAlive = false
	if err = DeleteRepo(client, fDB, args[0], clientKeepAlive); err != nil {
		log.Fatalf("Can't delete chart repository %s from database: %v", args[0], err, clientKeepAlive)
	}

	log.Infof("Successfully deleted the chart repository %s from database", args[0])
}
