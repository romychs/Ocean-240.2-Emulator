#! /usr/bin/python3
file_path = 'Disketa.okd'
sector = 1
try:
    with open(file_path, 'rb') as fi:
        # Read the entire content of the file
        while True:
            # Read 4 bytes at a time
            byte_chunk = fi.read(128)
            if not byte_chunk:
                break
            sn = str(sector).zfill(5)
            with open("sector-" + sn +'.dat', 'wb') as fo:
                fo.write(byte_chunk)
                fo.close()
            sector += 1
    fi.close()
except FileNotFoundError:
    print(f"Error: The file '{file_path}' was not found.")
except Exception as e:
    print(f"An error occurred: {e}")
