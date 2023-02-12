// Copyright (c) HashiCorp, Inc
// SPDX-License-Identifier: MPL-2.0
import { Construct } from "constructs";
import { App, TerraformStack } from "cdktf";
import * as google from '@cdktf/provider-google';

const project = 'qiita-cloud-run-grpc-acl';
const region = 'us-central1';
const repository = 'qiita-cloud-run-grpc-acl';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.provider.GoogleProvider(this, 'google', {
      project,
    });

    new google.artifactRegistryRepository.ArtifactRegistryRepository(this, 'dockerRegistry', {
      format: 'docker',
      location: region,
      repositoryId: 'docker-registry',
    });

    new google.cloudbuildTrigger.CloudbuildTrigger(this, 'buildTrigger', {
      filename: 'cloudbuild.yaml',
      github: {
        owner: 'hsmtkk',
        name: repository,
        push: {
          branch: 'main',
        },
      },
    });

    const locationProviderSA = new google.serviceAccount.ServiceAccount(this, 'locationProviderSA', {
      accountId: 'location-provider-sa',
    });

    const mapRenderSA = new google.serviceAccount.ServiceAccount(this, 'mapRenderSA', {
      accountId: 'map-render-sa',
    });

    new google.projectIamMember.ProjectIamMember(this, 'mapRenderSACanAccessSecret', {
      member: `serviceAccount:${mapRenderSA.email}`,
      project,
      role: 'roles/secretmanager.secretAccessor',
    });

    const googleMapAPIKey = new google.secretManagerSecret.SecretManagerSecret(this, 'googleMapAPIKey', {
      secretId: 'googleMapAPIKey',
      replication: {
        automatic: true,
      },
    });

    new google.secretManagerSecretVersion.SecretManagerSecretVersion(this, 'googleMapAPIKeyVersion', {
      secret: googleMapAPIKey.id,
      secretData: 'dummy',
    });

    const cloudRunNoAuth = new google.dataGoogleIamPolicy.DataGoogleIamPolicy(this, 'cloudRunNoAuth', {
      binding: [{
        role: 'roles/run.invoker',
        members: ['allUsers'],
      }],
    });

    const locationProvider = new google.cloudRunV2Service.CloudRunV2Service(this, 'locationProvider', {
      ingress: 'INGRESS_TRAFFIC_ALL',
      location: region,
      name: 'location-provider',
      template: {
        containers: [{
          image: 'us-central1-docker.pkg.dev/qiita-cloud-run-grpc-acl/docker-registry/location-provider:latest',
        }],        
        scaling: {
          minInstanceCount: 0,
          maxInstanceCount: 1,
        },
        serviceAccount: locationProviderSA.email,
      },
    });

    new google.cloudRunServiceIamPolicy.CloudRunServiceIamPolicy(this, 'locationProviderNoAuth', {
      location: region,
      policyData: cloudRunNoAuth.policyData,
      service: locationProvider.name,
    });

    const mapRender = new google.cloudRunV2Service.CloudRunV2Service(this, 'mapRender', {
      ingress: 'INGRESS_TRAFFIC_ALL',
      location: region,
      name: 'map-render',
      template: {
        containers: [{
          env: [{
            name: 'GOOGLE_MAP_API_KEY',
            valueSource: {
              secretKeyRef: {
                secret: googleMapAPIKey.secretId,
                version: 'latest',
              },
            },
          },
          {
            name: 'LOCATION_PROVIDER_URI',
            value: locationProvider.uri,
          },
          ],
          image: 'us-central1-docker.pkg.dev/qiita-cloud-run-grpc-acl/docker-registry/map-render:latest',
        }],        
        scaling: {
          minInstanceCount: 0,
          maxInstanceCount: 1,
        },
        serviceAccount: mapRenderSA.email,
      },
    });

    new google.cloudRunServiceIamPolicy.CloudRunServiceIamPolicy(this, 'mapRenderNoAuth', {
      location: region,
      policyData: cloudRunNoAuth.policyData,
      service: mapRender.name,
    });

  }
}

const app = new App();
new MyStack(app, "qiita-cloud-run-grpc-acl");
app.synth();
