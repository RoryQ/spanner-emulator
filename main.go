package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"github.com/googleapis/gax-go/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	ctx := context.Background()
	inst := os.Getenv("SPANNER_INSTANCE_ID")
	proj := os.Getenv("SPANNER_PROJECT_ID")
	db := os.Getenv("SPANNER_DATABASE_ID")
	dbs := os.Getenv("DATABASES")
	databases := resolveDBs(proj, inst, db, dbs)
	go ensureDatabases(ctx, databases)
	cmd := exec.Command("./gateway_main", "--hostname", "0.0.0.0")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func ensureDatabases(ctx context.Context, databases []dbase) error {
	if len(databases) == 0 {
		return nil
	}

	// clients
	ic, err := instance.NewInstanceAdminClient(ctx,
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithEndpoint("0.0.0.0:9010"),
	)
	if err != nil {
		return err
	}
	defer ic.Close()
	dc, err := database.NewDatabaseAdminClient(ctx,
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithEndpoint("0.0.0.0:9010"),
	)
	if err != nil {
		return err
	}
	defer dc.Close()

	errg, errctx := errgroup.WithContext(ctx)
	for _, dbase := range databases {
		errg.Go(func() error {
			return ensureDatabase(errctx, ic, dc, dbase)
		})
	}
	return errg.Wait()
}

func ensureDatabase(ctx context.Context, ic *instance.InstanceAdminClient, dc *database.DatabaseAdminClient, db dbase) error {
	if db.inst != "" && db.proj != "" {
		cir := &instancepb.CreateInstanceRequest{
			InstanceId: db.inst,
			Instance: &instancepb.Instance{
				Config:      "emulator-config",
				DisplayName: "",
				NodeCount:   1,
			},
			Parent: "projects/" + db.proj,
		}

		log.Printf("attempting to create instance %v\n", db.inst)
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

	if db.db != "" {
		log.Printf("attempting to create database %v\n", db)
		cdr := &databasepb.CreateDatabaseRequest{
			Parent:          "projects/" + db.proj + "/instances/" + db.inst,
			CreateStatement: "CREATE DATABASE `" + db.db + "`",
		}
		if cdrOp, err := dc.CreateDatabase(ctx, cdr); err != nil {
			// get the status code
			if errStatus, ok := status.FromError(err); ok {
				// if the resource already exists, continue
				if errStatus.Code() == codes.AlreadyExists {
					log.Printf("database already exists, continuing\n")
				} else {
					return err
				}
			} else {
				return err
			}
		} else {
			_, err = cdrOp.Wait(ctx)
			if err != nil {
				return err
			}
			log.Println("database created")
		}
	}

	return nil
}

var dbRegex = regexp.MustCompile(`^projects/([^/]+)/instances/([^/]+)(?:/databases/([^/]+))?$`)

func resolveDBs(proj, inst, db, dbs string) []dbase {
	result := []dbase{}
	if proj != "" && inst != "" {
		result = append(result, dbase{proj, inst, db})
	}
	list := strings.Split(dbs, ",")
	for _, l := range list {
		if l == "" {
			continue
		}
		m := dbRegex.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		result = append(result, dbase{m[1], m[2], m[3]})
	}
	return result
}

type dbase struct {
	proj, inst, db string
}
