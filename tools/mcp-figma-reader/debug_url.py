import re

def parse_figma_url(url):
    file_key_match = re.search(r"figma\.com/(?:file|design)/([a-zA-Z0-9]+)", url)
    node_id_match = re.search(r"node-id=([a-zA-Z0-9%-]+)", url)

    res = {}
    if file_key_match:
        res["file_key"] = file_key_match.group(1)
    if node_id_match:
        node_id = node_id_match.group(1).replace("-", ":")
        res["node_id"] = node_id

    return res

url = "https://www.figma.com/design/kQtFAU9yfZsqIEILCdS0Zg/-Seinetime-AI--UI-Design?node-id=403-23395&p=f&t=6mMnT9BLlfQ8IprP-0"
print(parse_figma_url(url))
