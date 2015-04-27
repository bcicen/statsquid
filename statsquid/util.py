from datetime import datetime

def format_bytes(b):
    b = float(b)
    if b < 1000:
        return '%i' % b + ' B'
    elif 1000 <= b < 1000000:
        return '%.1f' % float(b/1000) + ' KB'
    elif 1000000 <= b < 1000000000:
        return '%.1f' % float(b/1000000) + ' MB'
    elif 1000000000 <= b < 1000000000000:
        return '%.1f' % float(b/1000000000) + ' GB'
    elif 1000000000000 <= b:
        return '%.1f' % float(b/1000000000000) + ' TB'

def unix_time(dt):
    epoch = datetime.utcfromtimestamp(0)
    delta = dt - epoch
    return int(round(delta.total_seconds()))

def convert_type(value,to_type):
    if to_type == str:
        return str(value)
    if to_type == float:
        return float(value)
    if to_type == int:
        return int(value)

def output(msg):
    #wrapper to append date to printed message
    print('%s: %s' % (datetime.now().strftime('%Y-%m-%d %H:%M:%S'), msg))
