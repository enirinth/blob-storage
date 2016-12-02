import os
import numpy as np
import random
import string

numFiles = 10

def randFileName (fileLength):
    return ''.join(random.choice(string.lowercase) for i in range(fileLength))

def genFileNames(numFiles):
    file_names = []
    for i in range(numFiles):
        file_name_size = 10     # char in file name
        file_names.append(randFileName(file_name_size))
    return file_names

def genEverything(file_names):
    numFiles = len(file_names)

    out_log = "input.txt"

    fileSizeDist = np.random.zipf(2.0, numFiles)
    fileSizeDist = fileSizeDist * 10.0 / max(fileSizeDist)

    readReqDist = np.random.zipf(2.0, numFiles)

    with open(out_log, 'w') as outfile:
        for i in range(numFiles):
            newLine = file_names[i] + " " + str(fileSizeDist[i]) + " " + str(readReqDist[i]) + "\n"
            outfile.write(newLine)
    return


def main():
    # genFxn()

    file_names = genFileNames(numFiles)
    genEverything(file_names)

if __name__ == "__main__":
    main()
