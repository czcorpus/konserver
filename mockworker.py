import time
import json
import sys
import random
import os

class Random(object):

    def __init__(self, data):
        self._d = data
        self._i = 0

    def next(self):
        ans = self._d[self._i]
        self._i = (self._i + 1) % len(self._d)
        return ans

ran = Random([1, 2, 3, 4, 5, 6, 7, 2, 3, 4, 5, 1, 2, 4, 8, 1, 5, 2, 3, 7, 3, 2, 8])


def perform_task(command):
    time.sleep(ran.next())
    if command['fn'] == 'worker.conc_register':
        ans = dict(
            cachefile='/var/local/corpora/cache/syn2015/foobac.conc',
            already_running=False)
        return dict(
            status=2,
            error=None,
            result=ans
        )
    elif command['fn'] == 'worker.sum_and_repeat':
        args = command['args']
        return dict(status=2, error=None, result=[args['word']] * (args['a'] + args['b']))
    else:
        return dict(status=2, error=None, result=dict(some_data={}))


if __name__ == '__main__':
    ident = os.getpid()
    with open('/tmp/mockworker.txt', 'ab') as fw:
        fw.write('>>>>>>>>>>>>>>> INIT <<<<<<<<<<<<<<<<<<<<<<<\n')
        fw.flush()
        while True:
            command = sys.stdin.readline()
            fw.write('command: %s\n' % (command,))
            fw.flush()
            if command == '':
                break
            try:
                ans = perform_task(json.loads(command))
                fw.write('ANS: %s\n' % (ans,))
                sys.stdout.write(json.dumps(ans) + '\n')
            except Exception as ex:
                msg = '{0}: {1}'.format(ex.__class__.__name__, ex)
                fw.write('ANS: {0}\n'.format(json.dumps(dict(status=2, error=msg))))
                sys.stdout.write(json.dumps(dict(status=2, error=msg)) + '\n')
            fw.flush()
            sys.stdout.flush()



# python scripts/migration/to-0.12/update_conf.py -u 10 -p ./conf/config.xml