package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
)

var (
	debug   = flag.Bool("debug", true, "Enable debug logging")
	help    = flag.Bool("help", false, "Display usage information")
	version = flag.Bool("version", false, "Display version information")
)

// The CommitHash and Revision variables are set during building.
var (
	CommitHash = "<not set>"
	Revision   = "<not set>"
)

// Subscription describes a pull or push subscription
type Subscription struct {
	name string
	push string
}

// Topics describes a PubSub topic and its subscriptions.
type Topics map[string][]Subscription

func versionString() string {
	return fmt.Sprintf("pubsubc - build %s (%s) running on %s", Revision, CommitHash, runtime.Version())
}

// debugf prints debugging information.
func debugf(format string, params ...interface{}) {
	if *debug {
		fmt.Printf(format+"\n", params...)
	}
}

// fatalf prints an error to stderr and exits.
func fatalf(format string, params ...interface{}) {
	fmt.Fprintf(os.Stderr, os.Args[0]+": "+format+"\n", params...)
	os.Exit(1)
}

// create a connection to the PubSub service and create topics and subscriptions
// for the specified project ID.
func create(ctx context.Context, projectID string, topics Topics) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("Unable to create client to project %q: %s", projectID, err)
	}
	defer client.Close()

	fmt.Printf("Client connected with project ID %q", projectID)

	for topicID, subscriptions := range topics {
		fmt.Printf("  Creating topic %q\n", topicID)
		topic, err := client.CreateTopic(ctx, topicID)
		topic.EnableMessageOrdering = true

		if err != nil {
			return fmt.Errorf("Unable to create topic %q for project %q: %s", topicID, projectID, err)
		}

		for _, subscription := range subscriptions {
			if subscription.push != "" {
				fmt.Printf("    Creating PUSH subscription %q on topic %q to endpoint %q\n", subscription.name, topicID, subscription.push)
				_, err = client.CreateSubscription(ctx, subscription.name, pubsub.SubscriptionConfig{
					Topic: topic,
					PushConfig: pubsub.PushConfig{
						Endpoint: subscription.push,
					},
					AckDeadline:         300 * time.Second,
					ExpirationPolicy:    10 * time.Minute,
					RetainAckedMessages: false,
					RetryPolicy: &pubsub.RetryPolicy{
						MinimumBackoff: 1 * time.Second,
						MaximumBackoff: 5 * time.Second,
					},
					EnableMessageOrdering: true,
				})
				if err != nil {
					return fmt.Errorf("Unable to create subscription %q on topic %q for project %q: %s", subscription, topicID, projectID, err)
				}
			} else {
				fmt.Printf("    Creating PULL subscription %q on topic %q\n", subscription.name, topicID)
				_, err = client.CreateSubscription(ctx, subscription.name, pubsub.SubscriptionConfig{Topic: topic})
				if err != nil {
					return fmt.Errorf("Unable to create subscription %q on topic %q for project %q: %s", subscription, topicID, projectID, err)
				}
			}
		}
	}

	return nil
}

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Printf(`Usage: env PUBSUB_PROJECT1="project1\ntopic1\ntopic2:subscription1" %s`+"\n", os.Args[0])
		flag.PrintDefaults()
	}

	if *help {
		flag.Usage()
		return
	}

	if *version {
		fmt.Println(versionString())
		return
	}

	// Cycle over the numbered PUBSUB_PROJECT environment variables.
	for i := 1; ; i++ {
		// Fetch the enviroment variable. If it doesn't exist, break out.
		currentEnv := fmt.Sprintf("PUBSUB_PROJECT%d", i)
		env := os.Getenv(currentEnv)
		if env == "" {
			// If this is the first environment variable, print the usage info.
			if i == 1 {
				flag.Usage()
				os.Exit(1)
			}

			break
		} else {
			fmt.Printf("value of PUBSUB_PROJECT%d is %s", i, env)
		}

		// Separate the projectID from the topic definitions.
		parts := strings.Split(env, ",")
		if len(parts) < 2 {
			fatalf("%s: Expected at least 1 topic to be defined", parts)
		}

		projectId := parts[0]
		fmt.Printf("Project id is %s", projectId)

		// Separate the topicID from the subscription IDs.
		topics := make(Topics)
		for _, part := range parts[1:] {
			fmt.Printf("Part is %s", part)
			topicParts := strings.Split(part, ";")
			for _, subPart := range topicParts[1:] {
				subscriptionParts := strings.Split(subPart, "|")
				var sub = Subscription{
					name: subscriptionParts[0],
				}
				if len(subscriptionParts) == 2 {
					sub.push = subscriptionParts[1]
				}
				fmt.Printf("adding subscription to topic: topic = %s, sub = %s", topicParts[0], sub)
				topics[topicParts[0]] = append(topics[topicParts[0]], sub)
			}
		}

		for topicID, subscriptions := range topics {
			fmt.Printf("Project: %s\n\tTopic: %s\n\t\tSubscriptions:%s\n", projectId, topicID, subscriptions)
		}

		// Create the project and all its topics and subscriptions.
		if err := create(context.Background(), projectId, topics); err != nil {
			fatalf(err.Error())
		}
	}
}
