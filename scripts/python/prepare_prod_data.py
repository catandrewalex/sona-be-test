import argparse
import csv
from datetime import datetime
import json
import os
import string
from typing import List


FOLDER_NAME_OUT = 'out'


def read_csv_file(path: str) -> List[List[str]]:
    with open(path, 'r', newline='') as f_in:
        data = csv.reader(f_in, delimiter=',')
        return list(map(lambda row: list(map(lambda val: val.strip(), row)), data))[1:] # ignore header, so we start from idx 1

def _get_folderpath_filename_and_outfolderpath(path: str) -> [str, str, str]:
    folderpath = os.path.dirname(path)
    filename = os.path.splitext(os.path.basename(path))[0]
    outfolderpath = os.path.join(folderpath, 'out')
    return folderpath, filename, outfolderpath

if __name__ == '__main__':
    argparser = argparse.ArgumentParser()
    argparser.add_argument(
        '-s',
        '--student_data_file_path',
        type=str,
        default='data/csv/prod/students.csv',
    )
    argparser.add_argument(
        '-t',
        '--teacher_data_file_path',
        type=str,
        default='data/csv/prod/teachers.csv',
    )
    argparser.add_argument(
        '-tsf',
        '--teacher_special_fee_data_file_path',
        type=str,
        default='data/csv/prod/teacher_special_fees.csv',
    )
    argparser.add_argument(
        '-c',
        '--class_data_file_path',
        type=str,
        default='data/csv/prod/classes.csv',
    )
    argparser.add_argument(
        '-a',
        '--attendance_data_file_path',
        type=str,
        default='data/csv/prod/attendances.csv',
    )
    args = argparser.parse_args()

    if args.student_data_file_path:
        _, filename, outfolderpath = _get_folderpath_filename_and_outfolderpath(args.student_data_file_path)
        print(f'Students data are generated in: {outfolderpath}/{filename}.json')

        os.makedirs(outfolderpath, exist_ok=True)

        data = read_csv_file(args.student_data_file_path)
        deduplicator = dict() # Python 3.7's dict() and above are insertion-ordered
        email_deduplicator = dict()
        for idx, datum in enumerate(data[1:]):
            raw_name = string.capwords(datum[1].replace('  ', ' '))
            dob = datum[2]
            email = datum[5]
            instrument = datum[8]

            name, alias = '', ''
            if (idx := raw_name.find('(')) != -1:
                name, alias = raw_name[:idx].strip(), raw_name[idx:].translate(str.maketrans('', '', '()')).strip()
            else:
                name = raw_name
            first_name, *last_name = name.rsplit(' ', 1)
            last_name = last_name[0] if len(last_name) > 0 else ''
            username = '.'.join(name.lower().split()[:2])
            generated_password = datetime.strptime(dob, "%m/%d/%Y").date().strftime("%Y%m%d")

            if email_deduplicator.get(email) is not None:
                email_name, domain = email.split('@', 1)
                email = f'{email_name}+{username.split(".")[0]}@{domain}'

            deduplicator[f'{name}-{instrument}'] = {'name': name, 'alias': alias, 'dob': dob, 'instrument': instrument, 'email': email, 'username': username, 'password': generated_password, 'userDetail':{'firstName': first_name, 'lastName': last_name}, 'privilegeType': 200}
            email_deduplicator[email] = True

        clean_data = deduplicator.values()
        # This is only for debugging
        # with open(f'{outfolderpath}/{filename}.txt', 'w') as f_out:
        #     for datum in clean_data:
        #         f_out.write(f'{datum}\n')
            
        final_data = list(clean_data)
        for i in range(len(final_data)):
            del final_data[i]['name']
            del final_data[i]['alias']
            del final_data[i]['dob']
            del final_data[i]['instrument']
        with open(f'{outfolderpath}/{filename}.json', 'w') as f_out:
            json.dump({'data': final_data}, f_out, indent=4)

    # if args.teacher_data_file_path:
    #     _, filename, outfolderpath = _get_folderpath_filename_and_outfolderpath(args.teacher_data_file_path)
    #     print(f'Teachers data are generated in: {outfolderpath}/{filename}.json')

    #     os.makedirs(outfolderpath, exist_ok=True)

    #     data = read_csv_file(args.teacher_data_file_path)
    #     deduplicator = dict() # Python 3.7's dict() and above are insertion-ordered
    #     for idx, datum in enumerate(data[1:]):
    #         raw_name = string.capwords(datum[1].replace('  ', ' '))
    #         dob = datum[2]
    #         email = datum[5]
    #         instrument = datum[8]

    #         name, alias = '', ''
    #         if (idx := raw_name.find('(')) != -1:
    #             name, alias = raw_name[:idx].strip(), raw_name[idx:].translate(str.maketrans('', '', '()')).strip()
    #         else:
    #             name = raw_name
    #         first_name, *last_name = name.rsplit(' ', 1)
    #         last_name = last_name[0] if len(last_name) > 0 else ''
    #         username = '.'.join(name.lower().split()[:2])
    #         generated_password = datetime.strptime(dob, "%m/%d/%Y").date().strftime("%Y%m%d")

    #         deduplicator[f'{name}-{instrument}'] = {'name': name, 'alias': alias, 'dob': dob, 'instrument': instrument, 'email': email, 'username': username, 'password': generated_password, 'userDetail':{'firstName': first_name, 'lastName': last_name}, 'privilegeType': 200}

    #     clean_data = deduplicator.values()
    #     # This is only for debugging
    #     # with open(f'{outfolderpath}/{filename}.txt', 'w') as f_out:
    #     #     for datum in clean_data:
    #     #         f_out.write(f'{datum}\n')
            
    #     final_data = list(clean_data)
    #     for i in range(len(final_data)):
    #         del final_data[i]['name']
    #         del final_data[i]['alias']
    #         del final_data[i]['dob']
    #         del final_data[i]['instrument']
    #     with open(f'{outfolderpath}/{filename}.json', 'w') as f_out:
    #         json.dump({'data': final_data}, f_out, indent=4)
