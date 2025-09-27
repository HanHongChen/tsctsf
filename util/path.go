package util

import (
	"fmt"
	"net/url"
	"path"
	"github.com/HanHongChen/tsctsf/context"
	"github.com/HanHongChen/openapi-tsctsf/Npcf_PolicyAuthorization"
	"github.com/HanHongChen/openapi-tsctsf/Ntsnaf_BridgeInfoManagement"
	// "github.com/HanHongChen/bitbucket-openapi/Npcf_PolicyAuthorization"
	// "github.com/HanHongChen/bitbucket-openapi/Ntsnaf_BridgeInfoManagement"
)
//this place need to read config to get 
func GetNpcfPolicyAuthorizationClient(context *tsnContext.PCFContext) *Npcf_PolicyAuthorization.APIClient {
	configuration := Npcf_PolicyAuthorization.NewConfiguration()
	configuration.SetBasePath(context.tsnContext.PcfUri)
	client := Npcf_PolicyAuthorization.NewAPIClient(configuration)
	return client
}

// TODO: url??
func GetNtsnafBridgeInforManagementClient() *Ntsnaf_BridgeInfoManagement.APIClient {
	configuration := Ntsnaf_BridgeInfoManagement.NewConfiguration()
	configuration.SetBasePath("http://127.0.0.55:8000")
	client := Ntsnaf_BridgeInfoManagement.NewAPIClient(configuration)
	return client
}

func Split_appSessionId(Loc *url.URL) string {
	var temp string
	var slash int
	temp = Loc.String()
	for i := 0; i <= len(temp); i++ {
		if temp[i] == '/' {
			slash++
		}
		if slash == 6 {
			slash = i + 1
			break
		}
	}
	return temp[slash:]
}

func Split_appSessionId_str(Loc string) string {
	parsedUrl, err := url.Parse(Loc)
	if err != nil {
		panic(err)
	}

	lastSegment := path.Base(parsedUrl.Path)
	fmt.Println("App Session ID:", lastSegment)
	return lastSegment
}
