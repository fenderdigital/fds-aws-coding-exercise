try:
    # For local development
    from .routes import Router
    from .schemas.schemas import EventSchema
except:
    # For AWS Lambda deployment
    from routes import Router
    from schemas.schemas import EventSchema


def handler(event, context):
    event = EventSchema(**event)
    router = Router(event=event)

    return router.process_event()
