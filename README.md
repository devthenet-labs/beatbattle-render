# beatbattle-render

The render service of Beat Battle, a multiplayer 16-step drum-machine jam (contract:
[SPEC.md](../SPEC.md)). It turns a pattern into a WAV file: a deterministic, dependency-free drum synth behind one
stateless endpoint. Standard library only. It listens on `:8080` (override with `ADDR`) and writes nothing to disk.
Scaffolded by `maul init app` so that [maul](https://github.com/devthenet-labs/maul) can build intents in it and
preview its pull requests.

## HTTP API

Everything is served under `/render`; there is no path rewriting.

| Request | Response |
| --- | --- |
| `GET /render/healthz` | `200 {"status":"ok","revision":"<sha>"}` |
| `POST /render/wav` with a pattern JSON body | `200 audio/wav`, `Content-Disposition: inline; filename="beat.wav"` |

Errors are JSON `{"error":"<message>"}`: `400` for an invalid pattern, `405` for a wrong method, `404` for an unknown
path, `413` for a body over 16 KiB.

The pattern is the shared JSON (see the spec): `name` (1-40 characters, no control characters), `author` (1-24 of
`[A-Za-z0-9_-]`), integer `bpm` 60-200, integer `swing` 0-60 (delays every off-beat, the 2nd, 4th, 6th... sixteenth
note, by that percent of a step towards the next one) and `tracks` with exactly
`kick`, `snare` and `hat`, 16 booleans each. Unknown fields are rejected. `id`, `votes` and `createdAt` are accepted
and ignored.

```sh
curl -s -X POST localhost:8080/render/wav -o beat.wav -d '{"name":"Four","author":"me","bpm":120,"swing":0,
 "tracks":{"kick":[true,false,false,false,true,false,false,false,true,false,false,false,true,false,false,false],
 "snare":[false,false,false,false,true,false,false,false,false,false,false,false,true,false,false,false],
 "hat":[true,false,true,false,true,false,true,false,true,false,true,false,true,false,true,false]}}'
```

The output is 44100 Hz, mono, signed 16-bit PCM. A step is a sixteenth note (`60 / bpm / 4` s) and a pattern is 16
steps, so 120 BPM gives exactly 2.0 s: 88200 samples, a 176444-byte file (44-byte header). Voices still ringing at the
end are cut, never wrapped.

## Layout

- `synth/`: the pure core, no `net/http`. `Pattern` and `Validate`, `Render` (kick, snare and hat voices; noise from a
  seeded LCG, so output is bit-for-bit reproducible), `EncodeWAV` and `Onsets` (a windowed-RMS onset detector, used
  by tests now and acceptance checks later). Gains leave headroom so a full pattern does not clip.
- `server.go`, `main.go`: the thin HTTP adapter and process wiring.

## How it fits

`beatbattle-web` (`/`) serves the UI and `beatbattle-api` (`/api`) stores patterns. The browser composes the three
same-origin: the UI posts the current grid to `/render/wav` and plays the result. The services do not call each other.

## Develop

Use Go 1.26.6.

```sh
gofmt -l .
go vet ./...
go test -race ./...
ADDR=:8080 go run .
```

The root `Dockerfile` builds the runtime image a preview runs: a static binary on distroless, uid 65532, no shell,
listening on 8080. A preview runs it with a read-only root filesystem and no writable `/tmp`:

```sh
docker build -t beatbattle-render:local .
docker run --rm --read-only --cap-drop ALL --security-opt no-new-privileges \
  -p 127.0.0.1:8080:8080 beatbattle-render:local
curl -s localhost:8080/render/healthz
```

`.maul/Dockerfile` is a different image: the one maul's coding agent works in, which is maul's agent base plus the
pinned Go toolchain and this module's dependencies, offline. `.maul/agent.yaml` declares its immutable tag.

## CI and images

- `test` (`ci.yml`) tests the exact head and builds the runtime image without credentials.
- `agent image` (`agent-image.yml`) builds the agent image from `main`, also without credentials.
- `publish images` (`publish-images.yml`) hands each successful build to its trusted publisher, which pushes it to ECR.
  [`.github/actions/publish/README.md`](.github/actions/publish/README.md) describes the trust boundary.

### Repository variables

None of them is secret. Set them on this repository, never on the organization: an organization variable reaches every
repository with these workflows, a template repository and its copies included. maul's reference terraform module
(`deploy/terraform/aws/modules/app`) outputs the configuration as `github_variables_dotenv`; export it from your
terraform root and set it in one command:

```sh
terraform output -raw maul_app_beatbattle_render_variables | gh variable set -f - --repo devthenet-labs/beatbattle-render
```

Set the configuration first and the two `*_PUBLISH_ENABLED` gates last:

| Variable                   | Value                                                          | What it is                                                          |
| -------------------------- | -------------------------------------------------------------- | ------------------------------------------------------------------- |
| `PUBLISH_REPOSITORY_ID`    | `gh api repos/devthenet-labs/beatbattle-render --jq .id`       | this repository's immutable numeric ID, which the guard checks      |
| `PUBLISH_OWNER_ID`         | `gh api repos/devthenet-labs/beatbattle-render --jq .owner.id` | its owner's immutable numeric ID                                    |
| `AWS_REGION`               | `us-east-1`                                                    | the registry's AWS region                                           |
| `ECR_REGISTRY`             | `377946145366.dkr.ecr.us-east-1.amazonaws.com`                 | the ECR registry host                                               |
| `AGENT_IMAGE_REPOSITORY`   | `maul/app-envs/beatbattle-render`                              | the agent image's ECR repository, which `.maul/agent.yaml` names    |
| `AGENT_ROLE_ARN`           | the agent publisher role ARN                                   | assumed only by `publish-agent.yml`                                 |
| `AGENT_SIGNING_KEY`        | the signing key's `awskms:///` URI, or unset                   | optional: the KMS key the agent image is signed with                |
| `RUNTIME_IMAGE_REPOSITORY` | `maul/previews/beatbattle-render`                              | the runtime (preview) image's ECR repository                        |
| `RUNTIME_ROLE_ARN`         | the runtime publisher role ARN                                 | assumed only by `publish-runtime.yml`                               |
| `AGENT_PUBLISH_ENABLED`    | `true`                                                         | publishes the agent image; anything else skips it                   |
| `PREVIEW_PUBLISH_ENABLED`  | `true`                                                         | publishes runtime images; set it last, once previews are configured |

### Trusted workflows

Only these two workflows assume an AWS role, and each role trusts only its own workflow at `main`:

| Trusted workflow (`job_workflow_ref`)                                                    | Assumes            | Publishes                                                                            |
| ---------------------------------------------------------------------------------------- | ------------------ | ------------------------------------------------------------------------------------ |
| `devthenet-labs/beatbattle-render/.github/workflows/publish-runtime.yml@refs/heads/main` | `RUNTIME_ROLE_ARN` | the runtime image as `sha-<commit>`, and `main-<commit>` for a default-branch commit |
| `devthenet-labs/beatbattle-render/.github/workflows/publish-agent.yml@refs/heads/main`   | `AGENT_ROLE_ARN`   | the agent image as the `toolchain-v<N>` tag `.maul/agent.yaml` declares              |

### Changing the agent toolchain

Edit `.maul/Dockerfile` (or the dependencies in `go.mod`) and bump the `toolchain-v<N>` tag in `.maul/agent.yaml` in
the same commit: published tags are immutable, so a changed image needs a new tag.
