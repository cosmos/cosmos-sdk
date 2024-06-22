import pg from 'pg'

const {Client} = pg
const client = new Client({
    connectionString: import.meta.env.DATABASE_URL,
})
await client.connect()


export async function GET({params, request}) {
    const {query, variables, operationName} = params;

    return graphql(query, variables, operationName);
}

export async function POST({params, request}) {
    const {query, operationName, variables} = await request.json();
    return graphql(query, variables, operationName);
}

async function graphql(query, variables, operationName) {
    try {
        const res = await client.query(
            'select graphql.resolve($1, $2, $3);',
            [query, variables, operationName]
        );

        return new Response(
            JSON.stringify(res.rows[0].resolve),
            {
                headers: {
                    'Content-Type': 'application/json',
                }
            }
        )
    } catch (e) {
        return new Response(
            JSON.stringify({errors: [e.message]}),
            {
                status: 500,
                headers: {
                    'Content-Type': 'application/json',
                }
            }
        )
    }
}
