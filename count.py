with open("test") as file:
    line = file.readline()
    while line:
        sum0 = 0
        sum1 = 0
        sum2 = 0
        for i in range(3):
            line = file.readline().strip()
            line_split = line.split(' ')
            if line_split[0][-2:] == "ms":
                sum0 += float(line_split[0][:-2]) / 1000
            else:
                sum0 += float(line_split[0][:-1])
            sum1 += float(line_split[1].split('m')[0]) * 60 + float(line_split[1].split('m')[1][:-1])
            line = file.readline().strip()
            sum2 += float(line[:-1])
        line = file.readline()
        print(sum0/3, sum1/3, sum2/3)

