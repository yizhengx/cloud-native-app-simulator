/*
Copyright 2023 Telefonaktiebolaget LM Ericsson AB

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

package stressors

import (
	"application-emulator/src/client"
	model "application-model"
	"application-model/generated"
	"fmt"
	"net/http"
	"sync"
	"time"
	"math/rand"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var count int

func init() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Random Seed")
	// count = 0
}


// Headers to propagate from inbound to outbound
var incomingHeaders = []string{
	"User-Agent", "End-User", "X-Request-Id", "X-B3-TraceId", "X-B3-SpanId", "X-B3-ParentSpanId", "X-B3-Sampled", "X-B3-Flags",
}

// Extract relevant headers from the source request
func ExtractHeaders(request any) http.Header {
	// If this is a HTTP request, we should propagate the headers specified in incomingHeaders
	httpRequest, ok := request.(*http.Request)
	forwardHeaders := make(http.Header)

	if ok {
		for _, key := range incomingHeaders {
			if value := httpRequest.Header.Get(key); value != "" {
				forwardHeaders.Set(key, value)
			}
		}
	}

	// Override the content type
	forwardHeaders.Set("Content-Type", "application/json")

	return forwardHeaders
}

func httpRequest(service model.CalledService, forwardHeaders http.Header) generated.EndpointResponse {
	status, response, err :=
		client.POST(service.Service, service.Endpoint, service.Port, RandomPayload(service.RequestPayloadSize), forwardHeaders)

	if err != nil {
		return generated.EndpointResponse{
			Service:  &service,
			Status:   err.Error(),
			Protocol: "HTTP",
		}
	} else {
		return generated.EndpointResponse{
			Service:      &service,
			Status:       fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Protocol:     "HTTP",
			ResponseData: response,
		}
	}
}

func grpcRequest(service model.CalledService) generated.EndpointResponse {
	response, err :=
		client.GRPC(service.Service, service.Endpoint, service.Port, RandomPayload(service.RequestPayloadSize))

	if err != nil {
		return generated.EndpointResponse{
			Service:  &service,
			Status:   status.Convert(err).Code().String(),
			Protocol: "gRPC",
		}
	} else {
		return generated.EndpointResponse{
			Service:      &service,
			Status:       codes.OK.String(),
			Protocol:     "gRPC",
			ResponseData: response,
		}
	}
}

// Forward requests to all services sequentially and return REST or gRPC responses
func ForwardSequential(request any, services []model.CalledService) []generated.EndpointResponse {
	forwardHeaders := ExtractHeaders(request)
	length := 0
	dynamic := false
	for _, service := range services {
		if service.Probability != 0 {
			if !dynamic {
				dynamic = true
			}
			continue
		}
		length += service.TrafficForwardRatio
	}
	

	// Dynamic pattern: randomly pick one value
	picked_service := ""
	if dynamic {
		sum_prob := 0
		type s_p_pair struct {
			service_name string
			probability  int
			ratio int
		}
		var service_prob []s_p_pair
		for _, service := range services {
			if service.Probability != 0 {
				sum_prob += service.Probability
				service_prob = append(service_prob, s_p_pair{service.Service, sum_prob, service.TrafficForwardRatio})
			}
		}
		if (len(service_prob)==1) {
			sum_prob = 100
			service_prob = append(service_prob, s_p_pair{"not-existing-service", sum_prob, 0})
		}
		rand_value := rand.Intn(sum_prob-1) + 1
		// rand_value := (count)%100+10
		// count += 10
		for _, p := range service_prob {
			if rand_value <= p.probability {
				picked_service = p.service_name
				length += p.ratio
				break
			}
		}
	}

	responses := make([]generated.EndpointResponse, length, length)

	i := 0
	for _, service := range services {
		if service.Probability != 0 && picked_service != service.Service {
			continue
		}
		for j := 0; j < service.TrafficForwardRatio; j++ {
			if service.Protocol == "http" {
				response := httpRequest(service, forwardHeaders)
				responses[i] = response
			} else if service.Protocol == "grpc" {
				response := grpcRequest(service)
				responses[i] = response
			}
			i++
		}
	}

	return responses
}

func parallelHTTPRequest(responses []generated.EndpointResponse, i int, service model.CalledService, forwardHeaders http.Header, wg *sync.WaitGroup) {
	defer wg.Done()
	response := httpRequest(service, forwardHeaders)
	// No mutex needed since every response has its own index
	responses[i] = response
}

func parallelGRPCRequest(responses []generated.EndpointResponse, i int, service model.CalledService, wg *sync.WaitGroup) {
	defer wg.Done()
	response := grpcRequest(service)
	// No mutex needed since every response has its own index
	responses[i] = response
}

// Forward requests to all services in parallel using goroutines and return REST or gRPC responses
func ForwardParallel(request any, services []model.CalledService) []generated.EndpointResponse {
	forwardHeaders := ExtractHeaders(request)
	length := 0
	dynamic := false
	for _, service := range services {
		if service.Probability != 0 {
			if !dynamic {
				dynamic = true
			}
			continue
		}
		length += service.TrafficForwardRatio
	}

	// Dynamic pattern: randomly pick one value
	picked_service := ""
	if dynamic {
		sum_prob := 0
		type s_p_pair struct {
			service_name string
			probability  int
			ratio int
		}
		var service_prob []s_p_pair
		for _, service := range services {
			if service.Probability != 0 {
				sum_prob += service.Probability
				service_prob = append(service_prob, s_p_pair{service.Service, sum_prob, service.TrafficForwardRatio})
			}
		}
		if (len(service_prob)==1) {
			sum_prob = 100
			service_prob = append(service_prob, s_p_pair{"not-existing-service", sum_prob, 0})
		}
		rand_value := rand.Intn(sum_prob-1) + 1
		for _, p := range service_prob {
			if rand_value <= p.probability {
				picked_service = p.service_name
				length += p.ratio
				break
			}
		}
	}

	responses := make([]generated.EndpointResponse, length, length)
	wg := sync.WaitGroup{}

	i := 0
	for _, service := range services {
		if service.Probability != 0 && picked_service != service.Service {
			continue
		}
		for j := 0; j < service.TrafficForwardRatio; j++ {
			if service.Protocol == "http" {
				wg.Add(1)
				go parallelHTTPRequest(responses, i, service, forwardHeaders, &wg)
			} else if service.Protocol == "grpc" {
				wg.Add(1)
				go parallelGRPCRequest(responses, i, service, &wg)
			}
			i++
		}
	}

	wg.Wait()
	return responses
}
