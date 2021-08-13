// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/open-telemetry/opentelemetry-operator/api/v1alpha1"
	"github.com/open-telemetry/opentelemetry-operator/internal/config"
	. "github.com/open-telemetry/opentelemetry-operator/pkg/collector"
)

var logger = logf.Log.WithName("unit-tests")

func TestContainerNewDefault(t *testing.T) {
	// prepare
	otelcol := v1alpha1.OpenTelemetryCollector{}
	cfg := config.New(config.WithCollectorImage("default-image"))

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Equal(t, "default-image", c.Image)
}

func TestContainerWithImageOverridden(t *testing.T) {
	// prepare
	otelcol := v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{
			Image: "overridden-image",
		},
	}
	cfg := config.New(config.WithCollectorImage("default-image"))

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Equal(t, "overridden-image", c.Image)
}

func TestContainerConfigFlagIsIgnored(t *testing.T) {
	// prepare
	otelcol := v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{
			Args: map[string]string{
				"key":    "value",
				"config": "/some-custom-file.yaml",
			},
		},
	}
	cfg := config.New()

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Len(t, c.Args, 2)
	assert.Contains(t, c.Args, "--key=value")
	assert.NotContains(t, c.Args, "--config=/some-custom-file.yaml")
}

func TestContainerCustomVolumes(t *testing.T) {
	// prepare
	otelcol := v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{
			VolumeMounts: []corev1.VolumeMount{{
				Name: "custom-volume-mount",
			}},
		},
	}
	cfg := config.New()

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Len(t, c.VolumeMounts, 2)
	assert.Equal(t, "custom-volume-mount", c.VolumeMounts[1].Name)
}

func TestContainerCustomSecurityContext(t *testing.T) {
	// default config without security context
	c1 := Container(config.New(), logger, v1alpha1.OpenTelemetryCollector{Spec: v1alpha1.OpenTelemetryCollectorSpec{}})

	// verify
	assert.Nil(t, c1.SecurityContext)

	// prepare
	isPrivileged := true
	uid := int64(1234)

	// test
	c2 := Container(config.New(), logger, v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{
			SecurityContext: &corev1.SecurityContext{
				Privileged: &isPrivileged,
				RunAsUser:  &uid,
			},
		},
	})

	// verify
	assert.NotNil(t, c2.SecurityContext)
	assert.True(t, *c2.SecurityContext.Privileged)
	assert.Equal(t, *c2.SecurityContext.RunAsUser, uid)
}

func TestContainerEnvVarsOverridden(t *testing.T) {
	otelcol := v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{
			Env: []corev1.EnvVar{
				{
					Name:  "foo",
					Value: "bar",
				},
			},
		},
	}

	cfg := config.New()

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Len(t, c.Env, 1)
	assert.Equal(t, "foo", c.Env[0].Name)
	assert.Equal(t, "bar", c.Env[0].Value)
}

func TestContainerEmptyEnvVarsByDefault(t *testing.T) {
	otelcol := v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{},
	}

	cfg := config.New()

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Empty(t, c.Env)
}

func TestContainerResourceRequirements(t *testing.T) {
	otelcol := v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128M"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("256M"),
				},
			},
		},
	}

	cfg := config.New()

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Equal(t, resource.MustParse("100m"), *c.Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse("128M"), *c.Resources.Limits.Memory())
	assert.Equal(t, resource.MustParse("200m"), *c.Resources.Requests.Cpu())
	assert.Equal(t, resource.MustParse("256M"), *c.Resources.Requests.Memory())
}

func TestContainerDefaultResourceRequirements(t *testing.T) {
	otelcol := v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{},
	}

	cfg := config.New()

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Empty(t, c.Resources)
}

func TestContainerArgs(t *testing.T) {
	// prepare
	otelcol := v1alpha1.OpenTelemetryCollector{
		Spec: v1alpha1.OpenTelemetryCollectorSpec{
			Args: map[string]string{
				"metrics-level": "detailed",
				"log-level":     "debug",
			},
		},
	}
	cfg := config.New()

	// test
	c := Container(cfg, logger, otelcol)

	// verify
	assert.Contains(t, c.Args, "--metrics-level=detailed")
	assert.Contains(t, c.Args, "--log-level=debug")
}
