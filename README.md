tekton-slack-notify
====================

`tekton-slack-notify` is a tool for sending slack message into specified channel.
Unlike other slack plugins, this one supports message threadding so you don't spam
the channel with lots of updates about the same topic.

## Installation

To install this tool locally:

```console
$ go get github.com/ebay/tekton-slack-notify
```

Or install it as a Tekton task:

```console
$ kubectl apply -f https://raw.githubusercontent.com/eBay/tekton-slack-notify/refs/heads/main/task.yaml
```

## Usage

To run this tool locally:

```console
$ tekton-slack-notify -h
Usage of tekton-slack-notify:
      --channel string      channel ID to send message in
      --text string         message text
      --thread-ts string    thread ts, when specified, the message will be sent to thread
      --token-file string   file path containing slack API token
      --ts-file string      if specified, the message ts will be written into this file
```

Or run a TaskRun like below:

```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: slack-notify
spec:
  params:
  - name: channel
    value: <slack channel id>
  - name: text
    value: <text to send>
  - name: slack_secret
    value: <a secret with token key - an oauth slack token>
  - name: thread_ts
    value: <timestamp of parent message>
  taskRef:
    kind: Task
    name: slack-notify
```

## License

[Apache License 2.0][1]

[1]: https://www.apache.org/licenses/LICENSE-2.0
