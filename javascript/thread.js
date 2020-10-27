function packet()
{
	
}

function thread()
{
	this.type = "THREAD"

	this.smash = function(cdata){}

	this.Same = function(cdata)
	{
		if(this.type != cdata.type)
		{
			return false
		}

		return this.UUID == cdata.UUID
	}
}


