package main

import (
	"context"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
	"os/exec"
)

func main() {
	ctx := context.Background()
	go func() {
		if err := createInstance(ctx); err != nil {
			panic(err)}
	}()
	(&exec.Cmd{
		Path: "./gateway_main",
		Args: []string{"--hostname", "0.0.0.0"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}).Run()
}

func createInstance(ctx context.Context) error {
	inst := os.Getenv("SPANNER_INSTANCE_ID")
	proj := os.Getenv("SPANNER_PROJECT_ID")

	if inst != "" && proj != "" {
		ic, err := instance.NewInstanceAdminClient(ctx,
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithInsecure()),
			option.WithEndpoint("0.0.0.0:9010"),
		)
		if err != nil {
			return err
		}
		fmt.Println("connected")
		defer func() { _ = ic.Close() }()

		cir := &instancepb.CreateInstanceRequest{
			InstanceId: inst,
			Instance: &instancepb.Instance{
				Config:      "emulator-config",
				DisplayName: "",
				NodeCount:   1,
			},
			Parent: "projects/" + proj,
		}

		log.Printf("attempting to create instance %v\n", inst)
		if cirOp, err := ic.CreateInstance(ctx, cir, gax.WithGRPCOptions(grpc.WaitForReady(true))); err != nil {
			// get the status code
			if errStatus, ok := status.FromError(err); ok {
				// if the resource already exists, continue
				if errStatus.Code() == codes.AlreadyExists {
					log.Printf("instance already exists, continuing\n")
				} else {
					return err
				}
			} else {
				return err
			}
		} else {
			_, err = cirOp.Wait(ctx)
			if err != nil {
				return err
			}
			log.Println("instance created")
		}
	}

	return nil
}
