import sqlite3
import csv
import os

# Establish a connection to the SQLite database
conn = sqlite3.connect("emails.sqlite")
cursor = conn.cursor()

# Query the data from the 'emails' and 'deleted_emails' tables
cursor.execute("SELECT subject,labels FROM emails")
email_data = cursor.fetchall()

cursor.execute("SELECT subject,labels FROM deleted_emails")
deleted_email_data = cursor.fetchall()

# Combine the data from both tables
all_data = email_data + deleted_email_data

print("Total number of emails fetched:", len(all_data))

# Extract all unique labels
unique_labels = set()
for row in all_data:
    labels = row[1]
    labels = labels.split(', ')  # Convert the string of labels to a list
    unique_labels.update(labels)  # Convert labels to strings before adding to the set

# Create a label-to-integer mapping dictionary
label_to_int = {label: i for i, label in enumerate(sorted(unique_labels))}
num_labels = len(label_to_int)

print("Unique Label-to-integer mapping:", label_to_int)

# One-hot encode labels using the label-to-integer mapping
def one_hot_encode_labels(labels_str):
    labels = labels_str.split(', ')
    encoded_labels = [0] * num_labels
    for label in labels:  # Process labels as strings after splitting
        encoded_labels[label_to_int[label]] = 1
    return encoded_labels


print("One-hot encoded labels:", one_hot_encode_labels)

# Split the data into training, validation, and testing sets (70%, 15%, 15%)
train_data = all_data[:int(len(all_data) * 0.7)]
valid_data = all_data[int(len(all_data) * 0.7):int(len(all_data) * 0.85)]
test_data = all_data[int(len(all_data) * 0.85):]

# Function to write data to a CSV file
def write_data_to_csv(data, file_name):
    with open(file_name, "w", newline="", encoding="utf-8") as f:
        writer = csv.writer(f)
        writer.writerow(["text"] + list(label_to_int.keys()))  # Write the header row
        for row in data:
            subject = row[0]  # Assuming the email subject is the first column
            labels = row[1]  # Assuming the rest of the columns are labels
            encoded_labels = one_hot_encode_labels(labels)
            writer.writerow([subject] + encoded_labels)

# Write the data to the corresponding CSV files
write_data_to_csv(train_data, "train.csv")
write_data_to_csv(valid_data, "valid.csv")
write_data_to_csv(test_data, "test.csv")

# Close the SQLite connection
conn.close()

print("CSV files created successfully.")
