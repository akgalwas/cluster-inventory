# Runtime Migrator
The `runtime-migrator` application has the following tasks:
1. Connect to a Gardener project.
2. Retrieve all existing shoot specifications.
3. Migrate the shoot specs to the new Runtime custom resource (Runtime CRs created with this migrator have the `operator.kyma-project.io/created-by-migrator=true` label).
4. Saves the new Runtime CR to files.
5. Check if the new Runtime CR won't cause an update on Gardener.
6. Save the results of the comparison between the original shoot and the shoot KIM produces based on the new Runtime CR.
7. Apply the new Runtime CRs to the designated KCP cluster.
8. Save the migration results in the output json file.

## Build

In order to build the app, run the following command:

```bash
go build -o ./bin/runtime-migrator ./cmd/migration
``` 

## Usage

```bash
./runtime-migrator \
  -gardener-kubeconfig-path=/Users/myuser/gardener-kubeconfig.yml \
  -gardener-project-name=kyma-dev  \
  -kcp-kubeconfig-path=/Users/myuser/kcp-kubeconfig.yml \
  -output-path=/tmp/ \
  -dry-run=true \
  -input-file-path=input/runtimeIds.json \
  -input-type=json
```

The above **execution example** will: 
1. take the input from the `input/runtimeIds.json` file (json with runtime identifiers array)
1. proceed only with Runtime CRs creation for clusters listed in the input 
1. save output files in the `/tmp/<generated name>` directory. The output directory contains the following:
    - `migration-results.json` - the output file with the migration results
    - `runtimes` - the directory with the Runtime CRs files
    - `comparison-results` - the directory with the files generated during the comparison process
1. They will not be applied on the KCP cluster (`dry-run` mode)

The input can be also provided in the form of a text file:
```bash
./runtime-migrator \
  -gardener-kubeconfig-path=/Users/myuser/gardener-kubeconfig.yml \
  -gardener-project-name=kyma-stage  \
  -kcp-kubeconfig-path=/Users/myuser/kcp-kubeconfig.yml \
  -output-path=/tmp/ \
  -dry-run=true \
  -input-file-path=/Users/myuser/migrator-input/runtimeIds.txt \
  -input-type=txt
```

### Output example

```
2024/11/22 17:05:28 INFO Starting runtime-migrator
2024/11/22 17:05:28 gardener-kubeconfig-path: /Users/myuser/Downloads/kubeconfig-garden-kyma-stage.yaml
2024/11/22 17:05:28 kcp-kubeconfig-path: /Users/myuser/dev/config/sap
2024/11/22 17:05:28 gardener-project-name: kyma-stage
2024/11/22 17:05:28 output-path: /tmp/
2024/11/22 17:05:28 dry-run: true
2024/11/22 17:05:28 input-type: txt
2024/11/22 17:05:28 input-file-path: /Users/myuser/dev/source/infrastructure-manager/hack/runtime-migrator/input/runtimes-stage-docs.txt
2024/11/22 17:05:28
2024/11/22 17:05:33 INFO Migrating runtimes
2024/11/22 17:05:33 INFO Reading runtimeIds from input file
2024/11/22 17:05:43 INFO Runtime processed successfully (dry run) runtimeID=1df09b5b-0347-459d-aa0a-715db8fcaad7
2024/11/22 17:05:45 INFO Runtime processed successfully (dry run) runtimeID=ea439a5e-aa59-4e3e-8bfb-9bab1b31371e
2024/11/22 17:05:49 INFO Runtime processed successfully (dry run) runtimeID=d6eeafee-ffd5-4f23-97dc-a1df197b3b30
2024/11/22 17:05:52 WARN Runtime CR can cause unwanted update in Gardener runtimeID=99a38a99-e8d7-4b98-a6f2-5a54ed389c4d
2024/11/22 17:05:52 ERROR Failed to fetch shoot: shoot was deleted or the runtime ID is incorrect runtimeID=0a61a3c4-0ea8-4e39-860a-7853f0b6d180
2024/11/22 17:05:55 ERROR Failed to verify runtime runtimeID=6daf5f59-b0ab-44af-bb8e-7735fd609449
2024/11/22 17:05:55 INFO Migration completed. Successfully migrated runtimes: 3, Failed migrations: 2, Differences detected: 1
2024/11/22 17:05:55 INFO Migration results saved in: /tmp/migration-2024-11-22T17:05:33+01:00/migration-results.json
```

The migration results are saved in the `/tmp/migration-2024-11-22T17:05:33+01:00/migration-results.json` file.
The runtime custom resources are saved in the `/tmp/migration-2024-11-22T17:05:33+01:00/runtimes` directory.

The `migration-results.json` file contains the following content:
```json
[
   {
      "runtimeId": "1df09b5b-0347-459d-aa0a-715db8fcaad7",
      "shootName": "c-1228ddd",
      "status": "Success",
      "runtimeCRFilePath": "/tmp/migration-2024-11-22T17:05:33+01:00/runtimes/1df09b5b-0347-459d-aa0a-715db8fcaad7.yaml"
   },
   {
      "runtimeId": "ea439a5e-aa59-4e3e-8bfb-9bab1b31371e",
      "shootName": "c3a59d5",
      "status": "Success",
      "runtimeCRFilePath": "/tmp/migration-2024-11-22T17:05:33+01:00/runtimes/ea439a5e-aa59-4e3e-8bfb-9bab1b31371e.yaml"
   },
   {
      "runtimeId": "d6eeafee-ffd5-4f23-97dc-a1df197b3b30",
      "shootName": "c141da7",
      "status": "Success",
      "runtimeCRFilePath": "/tmp/migration-2024-11-22T17:05:33+01:00/runtimes/d6eeafee-ffd5-4f23-97dc-a1df197b3b30.yaml"
   },
   {
      "runtimeId": "99a38a99-e8d7-4b98-a6f2-5a54ed389c4d",
      "shootName": "c-71da0f2",
      "status": "ValidationDetectedUnwantedUpdate",
      "errorMessage": "Runtime may cause unwanted update in Gardener. Please verify the runtime CR.",
      "runtimeCRFilePath": "/tmp/migration-2024-11-22T17:05:33+01:00/runtimes/99a38a99-e8d7-4b98-a6f2-5a54ed389c4d.yaml",
      "comparisonResultDirPath": "/tmp/migration-2024-11-22T17:05:33+01:00/comparison-results/99a38a99-e8d7-4b98-a6f2-5a54ed389c4d"
   },
   {
      "runtimeId": "0a61a3c4-0ea8-4e39-860a-7853f0b6d180",
      "shootName": "",
      "status": "Error",
      "errorMessage": "Failed to fetch shoot: shoot was deleted or the runtime ID is incorrect"
   },
   {
      "runtimeId": "6daf5f59-b0ab-44af-bb8e-7735fd609449",
      "shootName": "c-1f810d0",
      "status": "ValidationError",
      "errorMessage": "Failed to verify runtime: audit logs configuration not found: missing region: 'australiaeast' for providerType: 'azure'",
      "runtimeCRFilePath": "/tmp/migration-2024-11-22T17:05:33+01:00/runtimes/6daf5f59-b0ab-44af-bb8e-7735fd609449.yaml"
   }
]

```
The following problems were detected in the above example:
- The runtime with the `0a61a3c4-0ea8-4e39-860a-7853f0b6d180` identifier was not found ; the identifier may be incorrect, or the corresponding shoot was deleted for some reason.
- The validation process for the runtime with the `6daf5f59-b0ab-44af-bb8e-7735fd609449` identifier failed. 
- The runtime with the `99a38a99-e8d7-4b98-a6f2-5a54ed389c4d` identifier may cause an unwanted update in the Gardener. The comparison results are saved in the `/tmp/migration-2024-11-22T17:05:33+01:00/comparison-results/99a38a99-e8d7-4b98-a6f2-5a54ed389c4d` directory.


The `/tmp/migration-2024-11-21T14:53:24+01:00/comparison-results/99a38a99-e8d7-4b98-a6f2-5a54ed389c4d` directory contains the following files:
- `c-71da0f2.diff`
- `converted-shoot.yaml`
- `original-shoot.yaml` 

The `c-71da0f2.diff` file contains the differences between the original shoot and the shoot that will be created based on the new Runtime CR. The `converted-shoot.yaml` file contains the shoot that will be created based on the new Runtime CR. The `original-shoot.yaml` file contains the shoot fetched from the Gardener.

## Configurable Parameters

This table lists the configurable parameters, their descriptions, and default values:

| Parameter | Description                                                                                                                                                                                                                                                                         | Default value       |
|-----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------|
| **kcp-kubeconfig-path** | Path to the Kubeconfig file of KCP cluster.                                                                                                                                                                                                                                         | `/path/to/kcp/kubeconfig` |
| **gardener-kubeconfig-path** | Path to the Kubeconfig file of Gardener cluster.                                                                                                                                                                                                                                    | `/path/to/gardener/kubeconfig` |
| **gardener-project-name** | Name of the Gardener project.                                                                                                                                                                                                                                                       | `gardener-project-name` |
| **output-path** | Path where generated report, and yamls will be saved. Directory has to exist.                                                                                                                                                                                                       | `/tmp/`             |
| **dry-run** | Dry-run flag. Has to be set to **false**, otherwise migrator will not apply the CRs on the KCP cluster.                                                                                                                                                                             | `true`              |
| **input-type** | Type of input to be used. Possible values: **txt** (will expect text file with one runtime identifier per line, [see the example](input/runtimeids_sample.txt)), and **json** (will expect `json` array with runtime identifiers, [see the example](input/runtimeids_sample.json)). | `json`              |
| **input-file-path** | Path to the file containing Runtimes to be migrated.                                                                                                                                                                                                                                | `/path/to/input/file`                    |

