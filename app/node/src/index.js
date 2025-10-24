
import { AppError } from './error.js';
import { getUserSubscription } from './getHandler.js'
import { httpWrapper } from './httpWrapper.js';
import { postHandler } from './postHandler.js'
const router = async (event) => {

    const method = (event.httpMethod || event.requestContext?.http?.method || "").toUpperCase();
    //routes
    switch(method){
        case 'GET':{
            const userId = event?.pathParameters?.userId
            if(!userId){
                throw new AppError(400, 'Missing UserId!')
            }
            //controller
            const response = await getUserSubscription(userId)
            return response
        }
        case 'POST':{
            const raw = event?.body ?? '';
            const isB64 = !!event?.isBase64Encoded;
            const text = isB64 ? Buffer.from(raw, 'base64').toString('utf8') : raw;

            const parsedBody =  JSON.parse(text || '{}')
            if(!parsedBody){
                throw new AppError(400, 'Missing body!')
            }
            //controller
            const result = await postHandler({eventType: parsedBody.eventType, parsedBody})
            return result
        }
        default:
            throw new AppError(405, 'Method not allowed!')
    }
};

export const handler = httpWrapper(router)
