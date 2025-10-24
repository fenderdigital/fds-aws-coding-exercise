import { AppError } from "./error.js";
import { createResponse } from "./utils.js";
export const httpWrapper = (handlerFunction)=>{
    return async (event)=>{
        try{
            const result = await handlerFunction(event)
            return createResponse(200, result)
        }catch(error){
            console.log(error)
            if(error instanceof(AppError)){
                return createResponse(error.status, {message: error.message}, )
            }else{
                return createResponse(500, {message: error.message})
            }
        }
    }
}