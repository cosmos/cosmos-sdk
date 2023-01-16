type PercentileBase = {
    at_ns: number;
    at_str: string;
};

type PercentilePerLatency = PercentileBase & { latency: number };
type PercentilePerBytesSize = PercentileBase & { size: number };
type PercentilePerCombined = PercentilePerBytesSize & PercentilePerLatency;

type PercentileRanking<T> = {
    p50: T,
    p75: T,
    p90: T,
    p95: T,
    p99: T,
};

type PerSecPoint = {
    sec: number;
    qps: number;
    bytes: number;
    bytes_rankings: PercentileRanking<PercentilePerBytesSize>,
    latency_rankings: PercentileRanking<PercentilePerLatency>,
};

type LoadOutputPoint = {
    start_time: string;
    avg_bytes_per_sec: number;
    avg_tx_per_sec: number;
    total_time: number;
    total_bytes: number;
    total_txs: number;
    p50: PercentilePerCombined;
    p75: PercentilePerCombined;
    p90: PercentilePerCombined;
    p95: PercentilePerCombined;
    p99: PercentilePerCombined;
    per_sec: PerSecPoint[];
};

export type LoadTestOutput = LoadOutputPoint[];

export const sampleOutput: LoadTestOutput = [
    {
        "start_time": "2022-12-17T23:34:32.843604-08:00",
        "avg_bytes_per_sec": 422.15393825983676,
        "avg_tx_per_sec": 105.53848456495919,
        "total_time": 18002911524,
        "total_bytes": 7600,
        "total_txs": 1900,
        "p50": {
            "at_ns": 12000740850,
            "at_str": "12.00074085s",
            "latency": 16942,
            "size": 4
        },
        "p75": {
            "at_ns": 2001010940,
            "at_str": "2.00101094s",
            "latency": 25159,
            "size": 4
        },
        "p90": {
            "at_ns": 15002840984,
            "at_str": "15.002840984s",
            "latency": 40838,
            "size": 4
        },
        "p95": {
            "at_ns": 16000796152,
            "at_str": "16.000796152s",
            "latency": 54770,
            "size": 4
        },
        "p99": {
            "at_ns": 3000639229,
            "at_str": "3.000639229s",
            "latency": 88433,
            "size": 4
        },
        "per_sec": [
            {
                "sec": 0,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 3025761,
                        "at_str": "3.025761ms",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 3126893,
                        "at_str": "3.126893ms",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 710362,
                        "at_str": "710.362µs",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 2010957,
                        "at_str": "2.010957ms",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 633480,
                        "at_str": "633.48µs",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 3025761,
                        "at_str": "3.025761ms",
                        "latency": 17800
                    },
                    "p75": {
                        "at_ns": 3126893,
                        "at_str": "3.126893ms",
                        "latency": 26296
                    },
                    "p90": {
                        "at_ns": 710362,
                        "at_str": "710.362µs",
                        "latency": 42240
                    },
                    "p95": {
                        "at_ns": 2010957,
                        "at_str": "2.010957ms",
                        "latency": 56953
                    },
                    "p99": {
                        "at_ns": 633480,
                        "at_str": "633.48µs",
                        "latency": 242929
                    }
                }
            },
            {
                "sec": 1,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 1002111655,
                        "at_str": "1.002111655s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 1002353024,
                        "at_str": "1.002353024s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 1002326247,
                        "at_str": "1.002326247s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 1002397596,
                        "at_str": "1.002397596s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 1001288678,
                        "at_str": "1.001288678s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 1002111655,
                        "at_str": "1.002111655s",
                        "latency": 14099
                    },
                    "p75": {
                        "at_ns": 1002353024,
                        "at_str": "1.002353024s",
                        "latency": 23419
                    },
                    "p90": {
                        "at_ns": 1002326247,
                        "at_str": "1.002326247s",
                        "latency": 39388
                    },
                    "p95": {
                        "at_ns": 1002397596,
                        "at_str": "1.002397596s",
                        "latency": 42895
                    },
                    "p99": {
                        "at_ns": 1001288678,
                        "at_str": "1.001288678s",
                        "latency": 81521
                    }
                }
            },
            {
                "sec": 2,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 2000905600,
                        "at_str": "2.0009056s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 2002236106,
                        "at_str": "2.002236106s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 2001414525,
                        "at_str": "2.001414525s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 2001683648,
                        "at_str": "2.001683648s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 2000836469,
                        "at_str": "2.000836469s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 2000905600,
                        "at_str": "2.0009056s",
                        "latency": 13925
                    },
                    "p75": {
                        "at_ns": 2002236106,
                        "at_str": "2.002236106s",
                        "latency": 24497
                    },
                    "p90": {
                        "at_ns": 2001414525,
                        "at_str": "2.001414525s",
                        "latency": 33965
                    },
                    "p95": {
                        "at_ns": 2001683648,
                        "at_str": "2.001683648s",
                        "latency": 53051
                    },
                    "p99": {
                        "at_ns": 2000836469,
                        "at_str": "2.000836469s",
                        "latency": 119294
                    }
                }
            },
            {
                "sec": 3,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 3001282673,
                        "at_str": "3.001282673s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 3002004828,
                        "at_str": "3.002004828s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 3001180017,
                        "at_str": "3.001180017s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 3000544530,
                        "at_str": "3.00054453s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 3000295758,
                        "at_str": "3.000295758s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 3001282673,
                        "at_str": "3.001282673s",
                        "latency": 19246
                    },
                    "p75": {
                        "at_ns": 3002004828,
                        "at_str": "3.002004828s",
                        "latency": 28618
                    },
                    "p90": {
                        "at_ns": 3001180017,
                        "at_str": "3.001180017s",
                        "latency": 40685
                    },
                    "p95": {
                        "at_ns": 3000544530,
                        "at_str": "3.00054453s",
                        "latency": 61919
                    },
                    "p99": {
                        "at_ns": 3000295758,
                        "at_str": "3.000295758s",
                        "latency": 144980
                    }
                }
            },
            {
                "sec": 4,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 4000650709,
                        "at_str": "4.000650709s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 4002067880,
                        "at_str": "4.00206788s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 4002361259,
                        "at_str": "4.002361259s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 4000745581,
                        "at_str": "4.000745581s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 4002262819,
                        "at_str": "4.002262819s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 4000650709,
                        "at_str": "4.000650709s",
                        "latency": 14782
                    },
                    "p75": {
                        "at_ns": 4002067880,
                        "at_str": "4.00206788s",
                        "latency": 21202
                    },
                    "p90": {
                        "at_ns": 4002361259,
                        "at_str": "4.002361259s",
                        "latency": 40556
                    },
                    "p95": {
                        "at_ns": 4000745581,
                        "at_str": "4.000745581s",
                        "latency": 62182
                    },
                    "p99": {
                        "at_ns": 4002262819,
                        "at_str": "4.002262819s",
                        "latency": 84402
                    }
                }
            },
            {
                "sec": 5,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 5001037936,
                        "at_str": "5.001037936s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 5000414399,
                        "at_str": "5.000414399s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 5001280145,
                        "at_str": "5.001280145s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 5002085141,
                        "at_str": "5.002085141s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 5000284963,
                        "at_str": "5.000284963s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 5001037936,
                        "at_str": "5.001037936s",
                        "latency": 14361
                    },
                    "p75": {
                        "at_ns": 5000414399,
                        "at_str": "5.000414399s",
                        "latency": 19172
                    },
                    "p90": {
                        "at_ns": 5001280145,
                        "at_str": "5.001280145s",
                        "latency": 24221
                    },
                    "p95": {
                        "at_ns": 5002085141,
                        "at_str": "5.002085141s",
                        "latency": 25801
                    },
                    "p99": {
                        "at_ns": 5000284963,
                        "at_str": "5.000284963s",
                        "latency": 58782
                    }
                }
            },
            {
                "sec": 6,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 6001696059,
                        "at_str": "6.001696059s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 6003318885,
                        "at_str": "6.003318885s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 6001420228,
                        "at_str": "6.001420228s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 6002873081,
                        "at_str": "6.002873081s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 6002339489,
                        "at_str": "6.002339489s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 6001696059,
                        "at_str": "6.001696059s",
                        "latency": 20756
                    },
                    "p75": {
                        "at_ns": 6003318885,
                        "at_str": "6.003318885s",
                        "latency": 29960
                    },
                    "p90": {
                        "at_ns": 6001420228,
                        "at_str": "6.001420228s",
                        "latency": 39011
                    },
                    "p95": {
                        "at_ns": 6002873081,
                        "at_str": "6.002873081s",
                        "latency": 57377
                    },
                    "p99": {
                        "at_ns": 6002339489,
                        "at_str": "6.002339489s",
                        "latency": 86379
                    }
                }
            },
            {
                "sec": 7,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 7002847341,
                        "at_str": "7.002847341s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 7003568297,
                        "at_str": "7.003568297s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 7002547576,
                        "at_str": "7.002547576s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 7002669274,
                        "at_str": "7.002669274s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 7001250358,
                        "at_str": "7.001250358s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 7002847341,
                        "at_str": "7.002847341s",
                        "latency": 16746
                    },
                    "p75": {
                        "at_ns": 7003568297,
                        "at_str": "7.003568297s",
                        "latency": 26185
                    },
                    "p90": {
                        "at_ns": 7002547576,
                        "at_str": "7.002547576s",
                        "latency": 41228
                    },
                    "p95": {
                        "at_ns": 7002669274,
                        "at_str": "7.002669274s",
                        "latency": 46180
                    },
                    "p99": {
                        "at_ns": 7001250358,
                        "at_str": "7.001250358s",
                        "latency": 83516
                    }
                }
            },
            {
                "sec": 8,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 8001049508,
                        "at_str": "8.001049508s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 8002710189,
                        "at_str": "8.002710189s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 8001463299,
                        "at_str": "8.001463299s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 8001421329,
                        "at_str": "8.001421329s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 8000696803,
                        "at_str": "8.000696803s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 8001049508,
                        "at_str": "8.001049508s",
                        "latency": 16531
                    },
                    "p75": {
                        "at_ns": 8002710189,
                        "at_str": "8.002710189s",
                        "latency": 22294
                    },
                    "p90": {
                        "at_ns": 8001463299,
                        "at_str": "8.001463299s",
                        "latency": 34387
                    },
                    "p95": {
                        "at_ns": 8001421329,
                        "at_str": "8.001421329s",
                        "latency": 51348
                    },
                    "p99": {
                        "at_ns": 8000696803,
                        "at_str": "8.000696803s",
                        "latency": 96810
                    }
                }
            },
            {
                "sec": 9,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 9003095175,
                        "at_str": "9.003095175s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 9001910994,
                        "at_str": "9.001910994s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 9001565881,
                        "at_str": "9.001565881s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 9000693512,
                        "at_str": "9.000693512s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 9000776907,
                        "at_str": "9.000776907s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 9003095175,
                        "at_str": "9.003095175s",
                        "latency": 18112
                    },
                    "p75": {
                        "at_ns": 9001910994,
                        "at_str": "9.001910994s",
                        "latency": 30880
                    },
                    "p90": {
                        "at_ns": 9001565881,
                        "at_str": "9.001565881s",
                        "latency": 46065
                    },
                    "p95": {
                        "at_ns": 9000693512,
                        "at_str": "9.000693512s",
                        "latency": 55382
                    },
                    "p99": {
                        "at_ns": 9000776907,
                        "at_str": "9.000776907s",
                        "latency": 77936
                    }
                }
            },
            {
                "sec": 10,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 10002608777,
                        "at_str": "10.002608777s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 10001927180,
                        "at_str": "10.00192718s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 10001462200,
                        "at_str": "10.0014622s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 10001956503,
                        "at_str": "10.001956503s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 10001515316,
                        "at_str": "10.001515316s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 10002608777,
                        "at_str": "10.002608777s",
                        "latency": 10920
                    },
                    "p75": {
                        "at_ns": 10001927180,
                        "at_str": "10.00192718s",
                        "latency": 19843
                    },
                    "p90": {
                        "at_ns": 10001462200,
                        "at_str": "10.0014622s",
                        "latency": 24228
                    },
                    "p95": {
                        "at_ns": 10001956503,
                        "at_str": "10.001956503s",
                        "latency": 27619
                    },
                    "p99": {
                        "at_ns": 10001515316,
                        "at_str": "10.001515316s",
                        "latency": 48557
                    }
                }
            },
            {
                "sec": 11,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 11002651883,
                        "at_str": "11.002651883s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 11002208627,
                        "at_str": "11.002208627s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 11000653322,
                        "at_str": "11.000653322s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 11000559840,
                        "at_str": "11.00055984s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 11001162516,
                        "at_str": "11.001162516s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 11002651883,
                        "at_str": "11.002651883s",
                        "latency": 21884
                    },
                    "p75": {
                        "at_ns": 11002208627,
                        "at_str": "11.002208627s",
                        "latency": 41220
                    },
                    "p90": {
                        "at_ns": 11000653322,
                        "at_str": "11.000653322s",
                        "latency": 63152
                    },
                    "p95": {
                        "at_ns": 11000559840,
                        "at_str": "11.00055984s",
                        "latency": 78941
                    },
                    "p99": {
                        "at_ns": 11001162516,
                        "at_str": "11.001162516s",
                        "latency": 151507
                    }
                }
            },
            {
                "sec": 12,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 12002151711,
                        "at_str": "12.002151711s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 12002314821,
                        "at_str": "12.002314821s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 12002508696,
                        "at_str": "12.002508696s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 12002472196,
                        "at_str": "12.002472196s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 12001039037,
                        "at_str": "12.001039037s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 12002151711,
                        "at_str": "12.002151711s",
                        "latency": 14251
                    },
                    "p75": {
                        "at_ns": 12002314821,
                        "at_str": "12.002314821s",
                        "latency": 18621
                    },
                    "p90": {
                        "at_ns": 12002508696,
                        "at_str": "12.002508696s",
                        "latency": 31462
                    },
                    "p95": {
                        "at_ns": 12002472196,
                        "at_str": "12.002472196s",
                        "latency": 52067
                    },
                    "p99": {
                        "at_ns": 12001039037,
                        "at_str": "12.001039037s",
                        "latency": 65697
                    }
                }
            },
            {
                "sec": 13,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 13002605904,
                        "at_str": "13.002605904s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 13000886493,
                        "at_str": "13.000886493s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 13001849600,
                        "at_str": "13.0018496s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 13001586231,
                        "at_str": "13.001586231s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 13000552063,
                        "at_str": "13.000552063s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 13002605904,
                        "at_str": "13.002605904s",
                        "latency": 22602
                    },
                    "p75": {
                        "at_ns": 13000886493,
                        "at_str": "13.000886493s",
                        "latency": 40154
                    },
                    "p90": {
                        "at_ns": 13001849600,
                        "at_str": "13.0018496s",
                        "latency": 61370
                    },
                    "p95": {
                        "at_ns": 13001586231,
                        "at_str": "13.001586231s",
                        "latency": 76776
                    },
                    "p99": {
                        "at_ns": 13000552063,
                        "at_str": "13.000552063s",
                        "latency": 204840
                    }
                }
            },
            {
                "sec": 14,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 14002856883,
                        "at_str": "14.002856883s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 14002577212,
                        "at_str": "14.002577212s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 14003304431,
                        "at_str": "14.003304431s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 14001565366,
                        "at_str": "14.001565366s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 14001251255,
                        "at_str": "14.001251255s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 14002856883,
                        "at_str": "14.002856883s",
                        "latency": 17402
                    },
                    "p75": {
                        "at_ns": 14002577212,
                        "at_str": "14.002577212s",
                        "latency": 23005
                    },
                    "p90": {
                        "at_ns": 14003304431,
                        "at_str": "14.003304431s",
                        "latency": 38277
                    },
                    "p95": {
                        "at_ns": 14001565366,
                        "at_str": "14.001565366s",
                        "latency": 47405
                    },
                    "p99": {
                        "at_ns": 14001251255,
                        "at_str": "14.001251255s",
                        "latency": 107938
                    }
                }
            },
            {
                "sec": 15,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 15003324318,
                        "at_str": "15.003324318s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 15002226500,
                        "at_str": "15.0022265s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 15001532917,
                        "at_str": "15.001532917s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 15002795584,
                        "at_str": "15.002795584s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 15000984647,
                        "at_str": "15.000984647s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 15003324318,
                        "at_str": "15.003324318s",
                        "latency": 18520
                    },
                    "p75": {
                        "at_ns": 15002226500,
                        "at_str": "15.0022265s",
                        "latency": 25903
                    },
                    "p90": {
                        "at_ns": 15001532917,
                        "at_str": "15.001532917s",
                        "latency": 44059
                    },
                    "p95": {
                        "at_ns": 15002795584,
                        "at_str": "15.002795584s",
                        "latency": 54403
                    },
                    "p99": {
                        "at_ns": 15000984647,
                        "at_str": "15.000984647s",
                        "latency": 79511
                    }
                }
            },
            {
                "sec": 16,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 16001569802,
                        "at_str": "16.001569802s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 16001261123,
                        "at_str": "16.001261123s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 16000902190,
                        "at_str": "16.00090219s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 16000796152,
                        "at_str": "16.000796152s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 16000387482,
                        "at_str": "16.000387482s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 16001569802,
                        "at_str": "16.001569802s",
                        "latency": 17337
                    },
                    "p75": {
                        "at_ns": 16001261123,
                        "at_str": "16.001261123s",
                        "latency": 25268
                    },
                    "p90": {
                        "at_ns": 16000902190,
                        "at_str": "16.00090219s",
                        "latency": 40710
                    },
                    "p95": {
                        "at_ns": 16000796152,
                        "at_str": "16.000796152s",
                        "latency": 54770
                    },
                    "p99": {
                        "at_ns": 16000387482,
                        "at_str": "16.000387482s",
                        "latency": 91260
                    }
                }
            },
            {
                "sec": 17,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 17000689296,
                        "at_str": "17.000689296s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 17001550140,
                        "at_str": "17.00155014s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 17002221923,
                        "at_str": "17.002221923s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 17000599323,
                        "at_str": "17.000599323s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 17001966637,
                        "at_str": "17.001966637s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 17000689296,
                        "at_str": "17.000689296s",
                        "latency": 13771
                    },
                    "p75": {
                        "at_ns": 17001550140,
                        "at_str": "17.00155014s",
                        "latency": 21627
                    },
                    "p90": {
                        "at_ns": 17002221923,
                        "at_str": "17.002221923s",
                        "latency": 50979
                    },
                    "p95": {
                        "at_ns": 17000599323,
                        "at_str": "17.000599323s",
                        "latency": 60682
                    },
                    "p99": {
                        "at_ns": 17001966637,
                        "at_str": "17.001966637s",
                        "latency": 104975
                    }
                }
            },
            { 
                "sec": 18,
                "qps": 100,
                "bytes": 400,
                "bytes_rankings": {
                    "p50": {
                        "at_ns": 18001763373,
                        "at_str": "18.001763373s",
                        "size": 4
                    },
                    "p75": {
                        "at_ns": 18002428799,
                        "at_str": "18.002428799s",
                        "size": 4
                    },
                    "p90": {
                        "at_ns": 18001972784,
                        "at_str": "18.001972784s",
                        "size": 4
                    },
                    "p95": {
                        "at_ns": 18001647182,
                        "at_str": "18.001647182s",
                        "size": 4
                    },
                    "p99": {
                        "at_ns": 18002554431,
                        "at_str": "18.002554431s",
                        "size": 4
                    }
                },
                "latency_rankings": {
                    "p50": {
                        "at_ns": 18001763373,
                        "at_str": "18.001763373s",
                        "latency": 17873
                    },
                    "p75": {
                        "at_ns": 18002428799,
                        "at_str": "18.002428799s",
                        "latency": 29114
                    },
                    "p90": {
                        "at_ns": 18001972784,
                        "at_str": "18.001972784s",
                        "latency": 43083
                    },
                    "p95": {
                        "at_ns": 18001647182,
                        "at_str": "18.001647182s",
                        "latency": 50769
                    },
                    "p99": {
                        "at_ns": 18002554431,
                        "at_str": "18.002554431s",
                        "latency": 71519
                    }
                }
            }
        ],
    }
]