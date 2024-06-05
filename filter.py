import sys, re

MULTI = re.compile(r'[\s]?"""\n?.*\n?"""\n?', re.MULTILINE)

if __name__=="__main__":
    filename = sys.argv[1]
    inf = open(filename).read()
    inf = MULTI.sub("", inf)
    open("output.graphql", "w").write(inf)
