# Tools

## update_versions.py

This script updates the versions of dappctrl:

* prod_data.sql
* dappinst/main.go

Only use at `release/*` branch.

### Usage

```
update_versions.py <DAPPCTRL_DIR>
```

Example of usage:

```
python update_versions.py ~/Projects/github.com/privatix/dappctrl
```

### Result

The result is updated files that described above.

Do not forget to commit the changes.
