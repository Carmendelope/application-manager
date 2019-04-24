/*
* Copyright (C) 2019 Nalej - All Rights Reserved
*/
package entities

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"testing"
)

func TestApplicationPackage(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Parameters package suite")
}