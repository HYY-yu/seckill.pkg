package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"os"
)

// 生成证书的小工具
// (X509证书)
func main() {
	name := os.Args[1]
	user := os.Args[2]

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	keyDer := x509.MarshalPKCS1PrivateKey(key)
	keyBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyDer,
	}

	keyFile, err := os.Create(name + "-key.pem")
	if err != nil {
		panic(err)
	}

	_ = pem.Encode(keyFile, &keyBlock)
	_ = keyFile.Close()

	commonName := user
	emailAddress := "feng@hyy-yu.space"
	org := "MyCo"
	orgUnit := "CO"
	city := "Hangzhou"
	state := "ZJ"
	country := "CN"

	subject := pkix.Name{
		CommonName:         commonName,
		Country:            []string{country},
		Locality:           []string{city},
		Organization:       []string{org},
		OrganizationalUnit: []string{orgUnit},
		Province:           []string{state},
	}

	asn1Obj, err := asn1.Marshal(subject.ToRDNSequence())
	if err != nil {
		panic(err)
	}

	csr := x509.CertificateRequest{
		RawSubject:         asn1Obj,
		EmailAddresses:     []string{emailAddress},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	ccrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csr, key)
	if err != nil {
		panic(err)
	}
	csrFile, err := os.Create(name + ".csr")
	if err != nil {
		panic(err)
	}
	_ = pem.Encode(csrFile, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: ccrBytes})
	_ = csrFile.Close()
}
