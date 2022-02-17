import json

fns = [
    "model_based_testing_traces_auto_action_outcome.json",
    "model_based_testing_traces_auto_queues.json",
    "model_based_testing_traces_P0.json",
    "model_based_testing_traces_P1.json",
    "model_based_testing_traces_P2.json",
    "model_based_testing_traces_P3.json",
    "model_based_testing_traces_P4.json",
    "model_based_testing_traces_P5.json",
    "model_based_testing_traces_P6.json",
    "model_based_testing_traces_P7.json",
    "model_based_testing_traces_P8.json",
]

for fn in fns:
    o = None
    with open(fn, "r") as fd:
        o = json.loads(fd.read())
    with open(fn, "w") as fd:
        fd.write(json.dumps(o, separators=(",", ":")))
