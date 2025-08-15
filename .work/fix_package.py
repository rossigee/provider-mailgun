import yaml; import sys
f = open(sys.argv[1], "r"); doc = yaml.safe_load(f); f.close()
doc.setdefault("spec", {})["controller"] = {"image": sys.argv[2]}
f = open(sys.argv[1], "w"); yaml.dump(doc, f, default_flow_style=False); f.close()
