def format_bytes(b):
    b = float(b)
    if b < 1000:
        return '%i' % b + 'b'
    elif 1000 <= b < 1000000:
        return '%.1f' % float(b/1000) + 'kb'
    elif 1000000 <= b < 1000000000:
        return '%.1f' % float(b/1000000) + 'mb'
    elif 1000000000 <= b < 1000000000000:
        return '%.1f' % float(b/1000000000) + 'gb'
    elif 1000000000000 <= b:
        return '%.1f' % float(b/1000000000000) + 'tb'
