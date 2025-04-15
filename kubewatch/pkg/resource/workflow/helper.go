/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package workflow

import (
	"crypto/tls"
	"github.com/devtron-labs/kubewatch/pkg/config"
	"github.com/go-resty/resty/v2"
	"log"
)

func publishEventsOnRest(jsonBody []byte, topic string, externalCdConfig *config.ExternalConfig) error {
	request := &publishRequest{
		Topic:   topic,
		Payload: jsonBody,
	}
	client := resty.New().SetDebug(true)
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resp, err := client.SetRetryCount(4).R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		SetAuthToken(externalCdConfig.Token).
		//SetResult().    // or SetResult(AuthSuccess{}).
		Post(externalCdConfig.ListenerUrl)

	if err != nil {
		log.Println("err in publishing over rest", "token ", externalCdConfig.Token, "body", request, err)
		return err
	}
	log.Println("res ", string(resp.Body()))
	return nil
}
