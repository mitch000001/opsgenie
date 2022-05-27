# Opsgenie cli

## How to use

Once compiled you have to set the API key either in an env var called `OPS_APIKEY` or create a config file in YAML format with the following content:

```yaml
apikey: YOUR_API_KEY
```

This config needs to be stored either in the current working directory (as config.yaml) or in your home folder below a directory called `.opsgenie`.