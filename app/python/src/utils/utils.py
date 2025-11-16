from datetime import datetime, timezone


def parse_iso8601(date_string: str) -> datetime:
    """
    Parse an ISO 8601 date string and return a timezone-aware datetime object.
    """
    return datetime.fromisoformat(date_string.replace("Z", "+00:00")).astimezone(
        timezone.utc
    )


def format_iso8601(dt: datetime) -> str:
    """
    Format a timezone-aware datetime object as an ISO 8601 date string.
    """
    return dt.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")
