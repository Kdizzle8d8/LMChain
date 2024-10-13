from datetime import datetime, timedelta

# Get today's date
today = datetime.now()

# Find the next Thursday
next_thursday = today + timedelta((3 - today.weekday()) % 7)

# Calculate the difference in days
days_until_thursday = (next_thursday - today).days

print(days_until_thursday)