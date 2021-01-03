import argparse, multiprocessing, random, os, shutil

CHUNK_SIZE = 1000000 / 8

class Workload:
    # file_name, file_size, file_conflict_ratio
    def __init__(self, fname, fsize, fconfrt):
        self.fname = fname
        self.fsize = fsize
        self.fconfrt = fconfrt

def data_gen(workload):
    limit = workload.fsize // workload.fconfrt
    print(workload.fname, workload.fsize, workload.fconfrt, limit)
    cnt = 0
    for i in range(workload.fsize):
        if i % CHUNK_SIZE == 0:
            # tx-y.csv
            file = open(workload.fname + '-' + str(cnt) + '.csv', 'a')
            cnt += 1
        a = random.randint(0, limit)
        b = random.randint(0, workload.fsize)
        file.write(str(a) + '\t' + str(b) + '\n')

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('--t1-size', type=int, default=1000000)
    parser.add_argument('--t1-conflict-ratio', type=int, default=100)
    parser.add_argument('--t2-size', type=int, default=1000000)
    parser.add_argument('--t2-conflict-ratio', type=int, default=100)
    args = parser.parse_args()

    folder = os.path.exists("data")
    if folder:
        shutil.rmtree("data")
    os.makedirs("data")

    workloads = [Workload("data/t1", args.t1_size, args.t1_conflict_ratio), Workload("data/t2", args.t2_size, args.t2_conflict_ratio)]
    with multiprocessing.Pool() as pool:
        pool.map(data_gen, workloads)