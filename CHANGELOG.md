# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial scaffold — OOD compute adapter for Amazon ECS Fargate, translating Open OnDemand job lifecycle calls to the Amazon ECS API.
- CLI commands: `submit` (JSON job spec from stdin → Fargate task, prints task ARN), `status <task-arn>` (OOD-normalized status), `delete <task-arn>` (stop a task), and `info <task-arn>` (full `DescribeTasks` response as JSON).
- Unit and substrate integration tests.

### Changed
- Upgraded to substrate v0.45.2 and removed the local `replace` directive; earlier upgraded to v0.44.3 (fixes substrate #241 ECS timestamp bug).
