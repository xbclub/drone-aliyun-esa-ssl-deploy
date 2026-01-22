package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	esa20240910 "github.com/alibabacloud-go/esa-20240910/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/zeromicro/go-zero/core/logx"
)

type Plugin struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	CertType        string
	PublicKey       string
	PrivateKey      string
	SiteId          int64
	client          *esa20240910.Client
}

func (p *Plugin) UploadSSL() error {
	request := &esa20240910.SetCertificateRequest{
		Type:        tea.String(p.CertType),
		Certificate: tea.String(p.PublicKey),
		PrivateKey:  tea.String(p.PrivateKey),
		SiteId:      tea.Int64(p.SiteId),
	}
	runtime := &util.RuntimeOptions{}

	resp, err := p.client.SetCertificateWithOptions(request, runtime)
	if err != nil {
		return wrapTeaError(err)
	}

	logx.Infof("Certificate uploaded successfully: %s", tea.StringValue(resp.Body.RequestId))
	return nil
}
func (p *Plugin) DeleteSSl(id string) error {
	request := &esa20240910.DeleteCertificateRequest{
		SiteId: tea.Int64(p.SiteId),
		Id:     tea.String(id),
	}
	runtime := &util.RuntimeOptions{}

	_, err := p.client.DeleteCertificateWithOptions(request, runtime)

	return err
}
func (p *Plugin) GetSSLList(name string) ([]*esa20240910.ListCertificatesResponseBodyResult, error) {
	request := &esa20240910.ListCertificatesRequest{
		SiteId:  tea.Int64(p.SiteId),
		Keyword: tea.String(name),
	}
	runtime := &util.RuntimeOptions{}

	resp, err := p.client.ListCertificatesWithOptions(request, runtime)
	if err != nil {
		return nil, wrapTeaError(err)
	}

	return resp.Body.Result, nil
}
func (p *Plugin) createClient() error {
	var cred credential.Credential
	var err error

	if p.AccessKeyId == "" && p.AccessKeySecret == "" {
		cred, err = credential.NewCredential(nil)
	} else if p.AccessKeyId != "" && p.AccessKeySecret != "" {
		cred, err = credential.NewCredential(&credential.Config{
			Type:            tea.String("access_key"),
			AccessKeyId:     tea.String(p.AccessKeyId),
			AccessKeySecret: tea.String(p.AccessKeySecret),
		})
	} else {
		return errors.New("both access key id and secret are required")
	}
	if err != nil {
		return err
	}

	cfg := &openapi.Config{
		Credential: cred,
		Endpoint:   tea.String(p.Endpoint),
	}
	p.client, err = esa20240910.NewClient(cfg)
	return err
}

func readCertificateFile(filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("certificate file path is empty")
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read certificate file '%s': %w", filePath, err)
	}
	return string(content), nil
}

func wrapTeaError(err error) error {
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) {
		msg := tea.StringValue(sdkErr.Message)
		if sdkErr.Data == nil {
			return errors.New(msg)
		}
		var data interface{}
		dec := json.NewDecoder(strings.NewReader(tea.StringValue(sdkErr.Data)))
		_ = dec.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			if recommend, ok := m["Recommend"]; ok {
				return fmt.Errorf("%s: %v", msg, recommend)
			}
		}
		return errors.New(msg)
	}
	return err
}

func main() {
	logconfig := logx.LogConf{
		Level:    "info",
		Mode:     "console",
		Encoding: "plain",
		Stat:     false,
	}
	logx.MustSetup(logconfig)
	publicKeyPath := os.Getenv("PLUGIN_PUBLIC_KEY")
	privateKeyPath := os.Getenv("PLUGIN_PRIVATE_KEY")
	domain := os.Getenv("PLUGIN_DOMAIN_NAME")
	if len(domain) == 0 {
		logx.Errorf("ssl domain name is empty")
		return
	}
	siteId := os.Getenv("PLUGIN_SITE_ID")
	publicKey, err := readCertificateFile(publicKeyPath)
	if err != nil {
		logx.Errorf("read public key: %v", err)
	}
	privateKey, err := readCertificateFile(privateKeyPath)
	if err != nil {
		logx.Errorf("read private key: %v", err)
	}

	endpoint := os.Getenv("PLUGIN_ENDPOINT")
	if endpoint == "" {
		endpoint = "esa.cn-hangzhou.aliyuncs.com"
	}
	certType := os.Getenv("PLUGIN_CERT_TYPE")
	if certType == "" {
		certType = "upload"
	}
	parseInt, err := strconv.ParseInt(siteId, 10, 64)
	if err != nil {
		logx.Errorf("parse site id: %v", err)
		return
	}
	plugin := &Plugin{
		AccessKeyId:     os.Getenv("PLUGIN_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("PLUGIN_ACCESS_KEY_SECRET"),
		Endpoint:        endpoint,
		CertType:        certType,
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
		SiteId:          parseInt,
	}
	err = plugin.createClient()
	if err != nil {
		logx.Errorf("create client failed: %v", err)
		return
	}
	list, err := plugin.GetSSLList("domain")
	if err != nil {
		logx.Errorf("get ssl list failed: %v", err)
		return
	}
	if len(list) > 0 {
		err = plugin.DeleteSSl(tea.StringValue(list[0].Id))
		if err != nil {
			logx.Errorf("delete ssl failed: %v", err)
			return
		}
	}
	if err := plugin.UploadSSL(); err != nil {
		log.Fatalf("upload certificate failed: %v", err)
	}
}
