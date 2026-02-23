package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc"
	pb "github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/lib"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/r2"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/mediasvc"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/postsvc"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/usersvc"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscredentials "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	mysql2 "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := loadEnv(); err != nil {
		log.Fatal("Error loading .env file")
	}

	//mysql
	db, err := NewMySQLConnection()
	if err != nil {
		log.Fatal("Error connecting to mysql")
	}

	r2Client, err := NewR2Client(context.Background())
	if err != nil {
		log.Fatal("Error connecting to r2")
	}

	// cloudflare r2
	var (
		mediaR2Repo   mediasvc.MediaR2Repository
		projectR2Repo projectsvc.ProjectR2Repository
	)
	{
		fmt.Println("---------R2 BUCKET NAME------------")
		bucketName := os.Getenv("R2_BUCKET_NAME")
		fmt.Println("--------------------------------")

		mediaR2Repo = r2.NewMediaR2Repository(r2Client, bucketName)
		projectR2Repo = r2.NewProjectR2Repository(r2Client, bucketName)
	}

	var (
		// repository
		mediaRepo   mediasvc.MediaRepository
		postRepo    postsvc.PostRepository
		projectRepo projectsvc.ProjectRepository
		userRepo    usersvc.UserRepository

		// search service
		mediaSvc   mediasvc.MediaService
		postSvc    postsvc.PostService
		projectSvc projectsvc.ProjectService
		userSvc    usersvc.UserService
	)
	{
		mediaRepo = mysql.NewMediaRepository(db, mediaR2Repo)
		postRepo = mysql.NewPostRepository(db)
		projectRepo = mysql.NewProjectRepository(db)
		userRepo = mysql.NewUserRepository(db)

		mediaSvc = mysql.NewMediaService(db, mediaR2Repo)
		postSvc = mysql.NewPostService(db)
		projectSvc = mysql.NewProjectService(db)
		userSvc = mysql.NewUserService(db)
	}

	var (
		// usecase
		mediaUcSvc   mediasvc.Service
		postUcSvc    postsvc.Service
		projectUcSvc projectsvc.Service
		userUcSvc    usersvc.Service
	)
	{
		mediaUcSvc = mediasvc.NewService(mediaRepo, mediaSvc)
		postUcSvc = postsvc.NewService(postRepo, postSvc, mediaRepo, mediaSvc, projectSvc, projectR2Repo)
		projectUcSvc = projectsvc.NewService(projectRepo, projectSvc, projectR2Repo)
		userUcSvc = usersvc.NewService(userRepo, userSvc)
	}

	var (
		// grpc
		mediaHandler   pb.MediaServiceServer
		postHandler    pb.PostServiceServer
		projectHandler pb.ProjectServiceServer
		userHandler    pb.UserServiceServer
	)
	{
		mediaHandler = grpc.NewMediaHandler(mediaUcSvc)
		postHandler = grpc.NewPostHandler(postUcSvc)
		projectHandler = grpc.NewProjectHandler(projectUcSvc)
		userHandler = grpc.NewUserHandler(userUcSvc)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("GRPC_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	tls := os.Getenv("GRPC_TLS_ENABLED") == "true"
	fmt.Println("tls: ", tls)
	var opts []grpc2.ServerOption

	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"

		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed loading certificates: %v\n", err)
		}

		opts = append(opts, grpc2.Creds(creds))
	}

	// opts = append(opts, grpc2.UnaryInterceptor(interceptor.AuthInterceptor))

	s := grpc2.NewServer(opts...)
	pb.RegisterMediaServiceServer(s, mediaHandler)
	pb.RegisterPostServiceServer(s, postHandler)
	pb.RegisterProjectServiceServer(s, projectHandler)
	pb.RegisterUserServiceServer(s, userHandler)

	reflection.Register(s)

	// Start gRPC
	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func NewMySQLConnection() (*sql.DB, error) {
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	if err != nil {
		log.Fatal("Error converting MYSQL_PORT to int")
	}

	sqlsec := mysql.Secret{
		Username: os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Engine:   "mysql",
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     port,
		DBName:   os.Getenv("MYSQL_DBNAME"),
	}
	sqlconf := &mysql2.Config{
		Net:       "tcp",
		ParseTime: true,
		Loc:       time.Local,
		Collation: "utf8mb4_general_ci",
	}
	sqlsec.Unmarshal(sqlconf)
	dsn := sqlconf.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error opening mysql connection")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging mysql connection")
	}

	return db, nil
}

func NewR2Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(awscredentials.NewStaticCredentialsProvider(os.Getenv("R2_ACCESS_KEY_ID"), os.Getenv("R2_ACCESS_KEY_SECRET"), "")),
		config.WithRegion("auto"), // Required by SDK but not used by R2
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", os.Getenv("R2_ACCOUNT_ID")))
	})

	return client, nil
}

func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}
