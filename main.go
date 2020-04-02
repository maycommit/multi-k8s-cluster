package main

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/eks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"log"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	awsConfig, sess, err := NewAwsConfig()

	g, err := token.NewGenerator(true, false)
	if err != nil {
		log.Fatalln(err)
		return
	}

	opts := &token.GetTokenOptions{
		ClusterID: aws.StringValue(awsConfig.Name),
		Session: sess,
	}

	t, err := g.GetWithOptions(opts)
	if err != nil {
		log.Fatalln("T: ", err)
		return
	}

	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(awsConfig.CertificateAuthority.Data))
	if err != nil {
		log.Println(err)
		return
	}

	namespace := "your-namespace"
	client, err := dynamic.NewForConfig(&rest.Config{
		Host: aws.StringValue(awsConfig.Endpoint),
		BearerToken: t.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	})
	if err != nil {
		panic(err)
	}

	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	list, err := client.Resource(deploymentRes).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	log.Println(list)

}

func NewToken() {

}

func NewAwsConfig() (*eks.Cluster, *session.Session, error) {
	id := ""
	secret := ""
	region := ""
	clusterName := ""

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials:                   credentials.NewStaticCredentials(id, secret, ""),
		MaxRetries:                    aws.Int(3),
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	}))

	_, err := sess.Config.Credentials.Get()
	if err != nil {
		log.Fatalln("ERROR: ", err)
		return nil, nil, err
	}

	cluster := eks.New(sess)

	res, err := cluster.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})

	if err != nil {
		return nil, nil, err
	}

	return res.Cluster, sess, nil
}
