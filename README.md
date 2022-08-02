# logwork

Logs work on a Jira-issue.

## Requirements

- `JIRA_API_TOKEN`: Env var containing your Jira-PAT

## Build

```bash
go install
asdf reshim golang # if using asdf
```

## Run

```bash
logwork -i QAB-258 -c "Daily scrum" -t 30
```
