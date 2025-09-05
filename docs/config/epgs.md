# EPGs

EPGs (Electronic Program Guides) define collections of TV program schedules from XML sources. Each EPG can contain
multiple sources.

## YAML Structure

```yaml
epgs:
  - name: epg-name
    source:
      - "http://example.com/epg-1.xml"
      - "/path/to/local/epg-2.xml.gz"
    proxy:
      enabled: true
```

## Fields

| Field    | Type                   | Required | Description                                                          |
|----------|------------------------|----------|----------------------------------------------------------------------|
| `name`   | `string`               | Yes      | Unique name identifier for this EPG                                  |
| `source` | `string` or `[]string` | Yes      | Array of EPG sources (URLs or file paths, XML format, .gz supported) |
| `proxy`  | [Proxy](./proxy.md)    | No       | EPG-specific proxy configuration                                     |

## Examples

### Basic EPG

```yaml
epgs:
  - name: tv-guide
    source:
      - "https://provider.com/guide.xml"
```

### Multi-Source EPG

```yaml
epgs:
  - name: combined-guide
    source:
      - "https://provider-1.com/epg.xml.gz"
      - "https://provider-2.com/schedule.xml"
      - "/local/custom-guide.xml"
```

### EPG with Proxy

```yaml
epgs:
  - name: international-guide
    source:
      - "https://international-provider.com/epg.xml"
    proxy:
      enabled: true
```
