//Multilingual message type definition
var dataType = {
	"int" : 1,
	"float" : 2,
	"char" : 3,
	"byte" : 4,
	"packet" : 5,
}



// A packet consists of all the data in a message.
// The packet is constructed by the data values and their length
//var packet = [
//	1, 		// version
//	2, 		// number of fields
//	dataType.int,	// first data type
//	1,		// length of the first data
//	dataType.char,	// second data type
//	12,		// length of the second data
//	42,		// first data begin
//	"Hello World!",	// second data begin. Notice the data length is 12
//]


function packet()
{
	this.version = 1
	this.data = []
	this.wrap = (function(val, type)
	{
		this.data.push({"type":type, "data":val})
	})

	this.encode = (function()
	{
		let types = []
		let content = []
		for(let i = 0; i < this.data.length; i++)
		{
			types.push(this.data[i].type)
			let length = this.data[i].data.length
			if(length == undefined) length = 1
			types.push(length)
			content.push(this.data[i].data)
		}

		let buffer = new ArrayBuffer(1+types.length+content.length)
		var index = 0
		buffer[index++] = this.version
		buffer[index++] = this.data.length

		for(let i = 0; i < types.length; i++)
		{
			buffer[index++] = types[i]
		}
		for(let i = 0; i < content.length; i++)
		{
			if(content[i].length != undefined)
			{
				for(let j = 0; j < content[i].length; j++)
				{
					buffer[index++] = content[i][j]
				}
			}
			else
			{
				buffer[index++] = content[i]
			}
		}
		return buffer
	})
}

function decodePacket(encPacket)
{
	if(encPacket.length <= 2)
	{
		return "bad packet"
	}
	
	var npacket = new packet()

	let index = 0
	npacket.version = encPacket[index++]
	numOfFields = encPacket[index++]

	for(let i = 0; i < numOfFields; i++)
	{
		npacket.data.push({"type":encPacket[index++], "length":encPacket[index++]})
	}

	for(let i = 0; i < numOfFields; i++)
	{
		let currentIndex = index
		let data = new ArrayBuffer(npacket.data[i].length)
		for(let j = 0; j < npacket.data[i].length; j++)
		{
			console.log(index, data[j], encPacket[index])
			data[j] = encPacket[index++]
		}

		npacket.data[i].data = cast(npacket.data[i].type, data)
	}

	return npacket
}

function cast(type, data)
{
	switch(type)
	{
		case dataType.int:
		{
			if(data.byteLength == 1)
			{
				return parseInt(data[0])
			}
			let ar = []
			for(let i = 0; i < data.byteLength; i++)
			{
				ar.push(parseInt(data[i]))
			}
			return ar
		}
		case dataType.float:
		{
			if(data.byteLength == 1)
			{
				return parseFloat(data[0])
			}
			let ar = []
			for(let i = 0; i < data.byteLength; i++)
			{
				ar.push(parseFloat(data[i]))
			}
			return ar
			
		}
		case dataType.char:
		{
			if(data.byteLength == 1)
			{
				return String(data[0])
			}
			let str = ""
			for(let i = 0; i < data.byteLength; i++)
			{
				str += String(data[i])
			}
			return str
			
		}
		case dataType.byte:
		return data
		case dataType.packet:
		return decodePacket(data)
	}
}

var p = new packet()
p.wrap(10.34, dataType.int)
let e = p.encode()
