# ood-fargate-adapter

OOD compute adapter for [Amazon ECS Fargate](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/AWS_Fargate.html). Translates Open OnDemand job lifecycle calls (submit / status / delete / info) to the Amazon ECS API.

## Commands

| Command | Description |
|---------|-------------|
| `submit` | Read a JSON job spec from stdin and run it as a Fargate task. Prints the task ARN on success. |
| `status <task-arn>` | Return OOD-normalised job status as JSON (`queued`, `running`, `completed`, `failed`, `cancelled`, `undetermined`). |
| `delete <task-arn>` | Stop a running Fargate task. |
| `info <task-arn>` | Print the full `DescribeTasks` API response as JSON. |

## Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--region` | `us-east-1` | AWS region |
| `--cluster` | _(empty)_ | ECS cluster ARN or name (can also be set per-job in the spec) |
| `--task-definition` | _(empty)_ | ECS task definition family:revision (can also be set per-job in the spec) |
| `--subnets` | _(empty)_ | Subnet IDs for the Fargate task, comma-separated (can also be set per-job in the spec) |
| `--security-groups` | _(empty)_ | Security group IDs, comma-separated |

## Job spec (stdin for `submit`)

```json
{
  "cluster": "arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster",
  "task_definition": "my-job-task:3",
  "script": "echo hello && sleep 10",
  "cpu": "512",
  "memory": "1024",
  "subnets": ["subnet-aabbccdd", "subnet-11223344"],
  "job_name": "my-ood-job-42",
  "env": {
    "MY_ENV_VAR": "value"
  }
}
```

`cpu` and `memory` follow ECS Fargate sizing units (CPU units and MiB). When omitted, the values from the task definition are used.

## Open OnDemand cluster YAML example

```yaml
# config/clusters.d/fargate.yml
v2:
  metadata:
    title: "Amazon ECS Fargate"
  login:
    host: "fargate.internal"
  job:
    adapter: "script"
    submit: "/usr/local/bin/ood-fargate-adapter submit --region us-east-1 --cluster my-cluster --task-definition my-job-task:3 --subnets subnet-aabbccdd"
    status: "/usr/local/bin/ood-fargate-adapter status --region us-east-1 --cluster my-cluster"
    delete: "/usr/local/bin/ood-fargate-adapter delete --region us-east-1 --cluster my-cluster"
    info:   "/usr/local/bin/ood-fargate-adapter info   --region us-east-1 --cluster my-cluster"
```

## Prerequisites

- Go 1.26+
- AWS credentials configured (environment variables, `~/.aws/credentials`, or IAM role)
- An existing ECS cluster and Fargate-compatible task definition
- VPC subnets with appropriate routing for Fargate tasks

## Build

```bash
go build -o ood-fargate-adapter .
```

## License

MIT — see [LICENSE](LICENSE).
