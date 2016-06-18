import json, os

with open("gtfs/stops.json", "r") as f:
    js = json.load(f)
    for i in range(len(js)):
        if "platform_code" in js[i]:
            js[i]["platform_code"] = str(js[i]["platform_code"])
    with open("gtfs/stops_clean.json", "w") as g:
        g.write(json.dumps(js))
try:
    os.remove("gtfs/stops_old.json")
except WindowsError:
    pass
os.rename("gtfs/stops.json", "gtfs/stops_old.json")
os.rename("gtfs/stops_clean.json", "gtfs/stops.json")