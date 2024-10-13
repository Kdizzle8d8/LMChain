from datetime import datetime, timedelta

# Get the current datetime
now = datetime.now()

# Calculate the number of days until the next Wednesday
days_until_wednesday = (2 - now.weekday() + 7) % 7

# If today is Wednesday, the next Wednesday is in 7 days
if days_until_wednesday == 0:
    days_until_wednesday = 7

# Print the result
print(days_until_wednesday)