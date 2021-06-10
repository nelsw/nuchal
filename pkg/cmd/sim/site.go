/*
 *
 * Copyright © 2021 Connor Van Elswyk ConnorVanElswyk@gmail.com
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package sim

import (
	"fmt"
	"github.com/nelsw/nuchal/pkg/util"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"strconv"
)

func newSite(simulations []simulation) error {

	if err := util.MakePath("html"); err != nil {
		return err
	}

	for _, simulation := range simulations {
		if err := newPage(simulation.productID, simulation.symbol(), "won", simulation.Won); err != nil {
			return err
		}
		if err := newPage(simulation.productID, simulation.symbol(), "lst", simulation.Lost); err != nil {
			return err
		}
		if err := newPage(simulation.productID, simulation.symbol(), "dnf", simulation.Trading); err != nil {
			return err
		}
	}

	log.Info().Msg(util.Sim + " .. ")
	log.Info().Msgf("%s ... charts ... http://localhost:%d", util.Sim, port())
	log.Info().Msg(util.Sim + " .. ")
	log.Info().Msg(util.Sim + " . ")

	fs := http.FileServer(http.Dir("html"))

	log.Print(http.ListenAndServe(fmt.Sprintf("localhost:%d", port()), logRequest(fs)))

	return nil
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msgf("%s ... %s %s %s", util.Sim, r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func port() int {
	if prt, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		return prt
	}
	return 8080
}
